"""
dynconf
~~~~~~~

This library allows to dynamically configure a service, e.g., some Django project.

"""
import logging

import etcd3
import six

__all__ = ('Config',)


class Config(object):
    """Provides an access to project's settings stored in etcd.

    Usage example::

        c = dynconf.Config(path='/configs/curiosity/')
        rover.set_velocity(
            c.integer(setting='velocity', default_value=5),
        )

    """
    def __init__(self, path, etcd=None, logger=None):
        """Sets up a new instance of the dynamic configuration.

        :param path: A path (prefix) to your project's settings.
            For example, project Curiosity might have settings such as velocity and is_camera_enabled.
            If the path is /configs/curiosity/, then the settings would be stored as the following etcd keys:
            /configs/curiosity/velocity and /configs/curiosity/is_camera_enabled.
        :param etcd: Optional etcd3 client configured to your taste,
            see https://python-etcd3.readthedocs.io/en/latest/usage.html.
            By default an etcd client connects to 127.0.0.1:2379 gRPC endpoint.
        :param logger: Optional logger configured to your taste.
            It helps to discover possible misconfigured settings.
            The logger works best when combined with JSON log formatter, see
            https://github.com/marselester/json-log-formatter.

        """
        self._cache = {}
        self._path = path
        self._path_len = len(path)

        self._etcd = etcd
        self._watch_id = None
        if self._etcd is None:
            self._etcd = etcd3.client()

        self._logger = logger
        if self._logger is None:
            self._logger = logging.getLogger('dynconf')

        self._load()

        try:
            self._watch_id = self._etcd.add_watch_prefix_callback(self._path, self._watch)
        except etcd3.exceptions.Etcd3Exception:
            self._logger.error('dynconf failed to watch settings', exc_info=True, extra={
                'path': self._path,
            })

    def _load(self):
        try:
            seq = self._etcd.get_prefix(self._path)
        except etcd3.exceptions.Etcd3Exception:
            self._logger.error('dynconf failed to load settings', exc_info=True, extra={
                'path': self._path,
            })
            return

        for value, meta in seq:
            k, v = self._decode_setting(meta.key, value, self._path_len)
            self._cache[k] = v

    def _watch(self, r):
        """Updates the settings cache when etcd keys change.

        This etcd callback is called when a etcd key for the given path (prefix)
        is created/updated/deleted.
        The keys and values (bytes) are decoded as utf-8 before being stored in the in-memory cache.

        """
        # When etcd shuts down, the _MultiThreadedRendezvous doesn't have events attribute.
        if not hasattr(r, 'events'):
            self._logger.error('dynconf watch failed: no events', extra={
                'path': self._path,
            })
            return

        # Perhaps the settings cache is empty because the etcd connection failed
        # when the Config instance was created.
        # Let's try loading the settings again.
        if not self._cache:
            self._load()

        for e in r.events:
            k, v = self._decode_setting(e.key, e.value, self._path_len)
            if isinstance(e, etcd3.events.PutEvent):
                self._cache[k] = v
            elif isinstance(e, etcd3.events.DeleteEvent):
                del self._cache[k]

    @staticmethod
    def _decode_setting(key, value, prefix_len):
        """Returns a decoded setting (key, value pair) from the etcd's raw key, value."""
        k = key[prefix_len:]
        k = k.decode('utf-8')
        v = value.decode('utf-8')
        return k, v

    def close(self):
        """Closes the underlying gRPC connection to etcd if it was established."""
        if not self._watch_id:
            return
        self._etcd.cancel_watch(self._watch_id)
        self._etcd.close()

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
