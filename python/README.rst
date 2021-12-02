=========================================
Dynamic configuration of a Python service
=========================================

This Python library provides a dynamic service configuration backed by etcd,
so there should be no need to redeploy a service to change its settings.

For example, project Curiosity expects settings such as `velocity = 10` and `is_camera_enabled = true`.
Let's save them in etcd with a path prefix `/configs/curiosity/`.

.. code-block:: console

    $ brew install etcd
    $ etcd
    $ etcdctl put /configs/curiosity/velocity 10
    OK
    $ etcdctl put /configs/curiosity/is_camera_enabled true
    OK

On the service side the settings are fetched from the same path.

.. code-block:: python

    import dynconf

    c = dynconf.Config(path='/configs/curiosity/')
    rover.set_velocity(
        c.integer(setting='velocity', default_value=5),
    )

Tests
-----

.. code-block:: console

    $ virtualenv venv
    $ source venv/bin/activate
    (venv) $ pip install -r requirements.txt
    (venv) $ tox
