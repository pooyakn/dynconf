import unittest

from dynconf import Config


class ConfigStringTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.c = Config('/configs/curiosity/')

    @classmethod
    def tearDownClass(cls):
        cls.c.close()

    def test_string_integer(self):
        self.c._cache = {'velocity': '10'}
        got = self.c.string('velocity', '5')
        self.assertEqual(got, '10')

    def test_string_name(self):
        self.c._cache = {'velocity': 'alice'}
        got = self.c.string('velocity', '5')
        self.assertEqual(got, 'alice')

    def test_none(self):
        self.c._cache = {'velocity': None}
        got = self.c.string('velocity', '5')
        self.assertEqual(got, '5')

    def test_int(self):
        self.c._cache = {'velocity': 100}
        got = self.c.string('velocity', '5')
        self.assertEqual(got, '5')

    def test_float(self):
        self.c._cache = {'velocity': 1.001}
        got = self.c.string('velocity', '5')
        self.assertEqual(got, '5')


class ConfigBooleanTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.c = Config('/configs/curiosity/')

    @classmethod
    def tearDownClass(cls):
        cls.c.close()

    def test_string_bool_true(self):
        self.c._cache = {'is_camera_enabled': 'true'}
        got = self.c.boolean('is_camera_enabled', False)
        self.assertTrue(got)

    def test_string_bool_false(self):
        self.c._cache = {'is_camera_enabled': 'false'}
        got = self.c.boolean('is_camera_enabled', False)
        self.assertFalse(got)

    def test_string_int(self):
        self.c._cache = {'is_camera_enabled': '10'}
        got = self.c.boolean('is_camera_enabled', False)
        self.assertFalse(got)

    def test_string_float(self):
        self.c._cache = {'is_camera_enabled': '10.001'}
        got = self.c.boolean('is_camera_enabled', False)
        self.assertFalse(got)

    def test_string_name(self):
        self.c._cache = {'is_camera_enabled': 'alice'}
        got = self.c.boolean('is_camera_enabled', False)
        self.assertFalse(got)

    def test_bytes(self):
        self.c._cache = {'is_camera_enabled': b'alice'}
        got = self.c.boolean('is_camera_enabled', False)
        self.assertFalse(got)

    def test_none(self):
        self.c._cache = {'is_camera_enabled': None}
        got = self.c.boolean('is_camera_enabled', False)
        self.assertFalse(got)

    def test_int(self):
        self.c._cache = {'is_camera_enabled': 100}
        got = self.c.boolean('is_camera_enabled', False)
        self.assertFalse(got)

    def test_float(self):
        self.c._cache = {'is_camera_enabled': 0.001}
        got = self.c.boolean('is_camera_enabled', False)
        self.assertFalse(got)


class ConfigIntegerTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.c = Config('/configs/curiosity/')

    @classmethod
    def tearDownClass(cls):
        cls.c.close()

    def test_string_integer(self):
        self.c._cache = {'velocity': '10'}
        got = self.c.integer('velocity', 5)
        self.assertEqual(got, 10)

    def test_string_name(self):
        self.c._cache = {'velocity': 'alice'}
        got = self.c.integer('velocity', 5)
        self.assertEqual(got, 5)

    def test_bytes(self):
        self.c._cache = {'velocity': b'alice'}
        got = self.c.integer('velocity', 5)
        self.assertEqual(got, 5)

    def test_none(self):
        self.c._cache = {'velocity': None}
        got = self.c.integer('velocity', 5)
        self.assertEqual(got, 5)

    def test_int(self):
        self.c._cache = {'velocity': 100}
        got = self.c.integer('velocity', 5)
        self.assertEqual(got, 100)

    def test_float(self):
        self.c._cache = {'velocity': 1.001}
        got = self.c.integer('velocity', 5)
        self.assertEqual(got, 1)
