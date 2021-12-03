"""
dynconf
~~~~~~~

This library allows to dynamically configure a service, e.g., some Django project.

"""
import time
import logging

import etcd3
import six
import dateutil.parser

__all__ = ('Config',)


class Config(object):
    """Provides an access to project's settings stored in etcd.

    Usage example:

    .. code-block:: python

        c = dynconf.Config(path='/configs/curiosity/')
        rover.set_velocity(
            c.integer(setting='velocity', default_value=5),
        )

    """
    def __init__(self, path, create_etcd3_client=None, logger=None):
        """Sets up a new instance of the dynamic configuration.

        :param path: A path (prefix) to your project's settings.
            For example, project Curiosity might have settings such as velocity and is_camera_enabled.
            If the path is /configs/curiosity/, then the settings would be stored as the following etcd keys:
            /configs/curiosity/velocity and /configs/curiosity/is_camera_enabled.
        :param create_etcd3_client: An optional function that must return etcd3 client configured to your taste,
            see https://python-etcd3.readthedocs.io/en/latest/usage.html.
            By default an etcd client connects to 127.0.0.1:2379 gRPC endpoint.
            Beware of exceptions if an etcd client is misconfigured.
            It's recommended to catch and log them
            to give an opportunity for the service (a Django project)
            to start and run using default settings.
        :param logger: Optional logger configured to your taste.
            It helps to discover possible misconfigured settings.
            The logger works best when combined with JSON log formatter, see
            https://github.com/marselester/json-log-formatter.

        """
        self._cache = {}
        self._path = path

        self._create_etcd3_client = create_etcd3_client
        if self._create_etcd3_client is None:
            self._create_etcd3_client = etcd3.client
        self._etcd = self._create_etcd3_client()

        self._logger = logger
        if self._logger is None:
            self._logger = logging.getLogger('dynconf')

        # Note, these functions shouldn't block for long to avoid interfering with service launch.
        self._watch_id = None
        self._load()
        self._create_watcher()

    def close(self):
        """Closes the underlying gRPC etcd connection if it was established."""
        # Non-empty watch_id of the keys watcher indicates that
        # there was an etcd connection at some point, so it should be cancelled.
        if self._watch_id is None:
            return
        self._etcd.cancel_watch(self._watch_id)
        self._etcd.close()
        self._watch_id = None

    def _load(self):
        """Fetches all the settings from etcd and populates the in-memory cache.

        Note, the etcd exceptions are logged instead of propagating,
        and a boolean status is returned to indicate success or failure.

        """
        try:
            seq = self._etcd.get_prefix(self._path)
        except etcd3.exceptions.Etcd3Exception:
            self._logger.error('dynconf failed to load settings', exc_info=True, extra={
                'path': self._path,
            })
            return False

        path_len = len(self._path)
        for value, meta in seq:
            k, v = self._decode_setting(meta.key, value, path_len)
            self._cache[k] = v

        return True

    def _create_watcher(self):
        """Creates an etcd key watcher that runs in a separate thread.

        Its ID is stored so the watcher can be cancelled when Config instance is discarded.
        The _watch callback is fired on any changes to etcd keys with given path prefix.

        Note, the etcd exceptions are logged instead of propagating,
        and a boolean status is returned to indicate success or failure.

        """
        try:
            self._watch_id = self._etcd.add_watch_prefix_callback(self._path, self._watch)
        except etcd3.exceptions.Etcd3Exception:
            self._logger.error('dynconf failed to watch settings', exc_info=True, extra={
                'path': self._path,
            })
            return False

        return True

    def _watch(self, r):
        """Updates the settings cache when etcd keys change.

        This etcd callback is called when a etcd key for the given path (prefix)
        is created/updated/deleted.
        The keys and values (bytes) are decoded as utf-8 before being stored in the in-memory cache.

        """
        if hasattr(r, 'events'):
            path_len = len(self._path)

            for e in r.events:
                k, v = self._decode_setting(e.key, e.value, path_len)
                if isinstance(e, etcd3.events.PutEvent):
                    self._cache[k] = v
                elif isinstance(e, etcd3.events.DeleteEvent):
                    del self._cache[k]

            return

        # When etcd server shuts down, the _MultiThreadedRendezvous doesn't have events attribute.
        # This fact is used to try establish a connection again, loading the settings,
        # and creating a new keys watcher.
        self._logger.error('dynconf watch failed: no events', extra={
            'path': self._path,
        })

        # Due to the connectivity issues a new etcd client has to be created,
        # because python-etcd3 doesn't support reconnections unlike the Go etcd client.
        # It seems there is no need to call self._etcd.close() here since gRPC connection failed.
        # When it is closed, there is an exception in logs: ValueError: Cannot invoke RPC: Channel closed!
        self._etcd = self._create_etcd3_client()

        # Keep trying until the settings are loaded and the watcher is created,
        # but give up when this Config instance is closed (the watcher ID is None).
        wait_seconds = 10
        while True:
            if self._watch_id is None:
                self._logger.info('dynconf give up reconnecting: Config closed', extra={
                    'path': self._path,
                })
                break

            self._logger.info('dynconf wait for {}s to reconnect'.format(wait_seconds), extra={
                'path': self._path,
            })
            time.sleep(wait_seconds)
            self._logger.info('dynconf reconnecting', extra={
                'path': self._path,
            })
            if not self._load():
                continue
            if self._create_watcher():
                self._logger.info('dynconf reconnected', extra={
                    'path': self._path,
                })
                break

    @staticmethod
    def _decode_setting(key, value, prefix_len):
        """Returns a decoded setting (key, value pair) from the etcd's raw key, value."""
        k = key[prefix_len:]
        k = k.decode('utf-8')
        v = value.decode('utf-8')
        return k, v

    def settings(self):
        """Returns all the project's settings as a dict where keys and values are strings."""
        return self._cache.copy()

    def string(self, setting, default_value):
        """Returns the string value of the given setting,
        or default_value if it wasn't found or type conversion failed.

        :param setting: A setting name, e.g., name.
        :param default_value: A default setting value, e.g., bob.

        """
        if setting not in self._cache:
            self._logger.error('dynconf setting not found: {}'.format(setting), extra={
                'path': self._path,
                'setting': setting,
            })
            return default_value

        v = self._cache[setting]
        if isinstance(v, six.string_types):
            return v

        self._logger.error('dynconf invalid string setting: {}'.format(setting), extra={
            'path': self._path,
            'setting': setting,
            'value': v,
        })
        return default_value

    def boolean(self, setting, default_value):
        """Returns the boolean value of the given setting,
        or default_value if it wasn't found or type conversion failed.

        :param setting: A setting name, e.g., is_camera_enabled.
        :param default_value: A default setting value, e.g., True.

        """
        if setting not in self._cache:
            self._logger.error('dynconf setting not found: {}'.format(setting), extra={
                'path': self._path,
                'setting': setting,
            })
            return default_value

        v = self._cache[setting]
        if isinstance(v, six.string_types) and v in ('true', 'false'):
            return v == 'true'

        self._logger.error('dynconf invalid boolean setting: {}'.format(setting), extra={
            'path': self._path,
            'setting': setting,
            'value': v,
        })
        return default_value

    def integer(self, setting, default_value):
        """Returns the integer value of the given setting,
        or default_value if it wasn't found or type conversion failed.

        :param setting: A setting name, e.g., velocity.
        :param default_value: A default setting value, e.g., 5.

        """
        if setting not in self._cache:
            self._logger.error('dynconf setting not found: {}'.format(setting), extra={
                'path': self._path,
                'setting': setting,
            })
            return default_value

        v = self._cache[setting]
        try:
            return int(v)
        except (ValueError, TypeError) as exc:
            self._logger.error('dynconf invalid integer setting: {}'.format(setting), exc_info=True, extra={
                'path': self._path,
                'setting': setting,
                'value': v,
            })
            return default_value

    def float(self, setting, default_value):
        """Returns the float value of the given setting,
        or default_value if it wasn't found or type conversion failed.

        :param setting: A setting name, e.g., temperature.
        :param default_value: A default setting value, e.g., 36.6.

        """
        if setting not in self._cache:
            self._logger.error('dynconf setting not found: {}'.format(setting), extra={
                'path': self._path,
                'setting': setting,
            })
            return default_value

        v = self._cache[setting]
        try:
            return float(v)
        except (ValueError, TypeError) as exc:
            self._logger.error('dynconf invalid float setting: {}'.format(setting), exc_info=True, extra={
                'path': self._path,
                'setting': setting,
                'value': v,
            })
            return default_value

    def date(self, setting, default_value):
        """Returns the date value of the given setting,
        or default_value if it wasn't found or type conversion failed.

        :param setting: A setting name, e.g., launched_at.
        :param default_value: A default setting value, e.g., 2021-11-30T20:14:05.134115+00:00.

        """
        if setting not in self._cache:
            self._logger.error('dynconf setting not found: {}'.format(setting), extra={
                'path': self._path,
                'setting': setting,
            })
            return default_value

        v = self._cache[setting]
        try:
            return dateutil.parser.parse(v)
        except (ValueError, TypeError) as exc:
            self._logger.error('dynconf invalid date setting: {}'.format(setting), exc_info=True, extra={
                'path': self._path,
                'setting': setting,
                'value': v,
            })
            return default_value
