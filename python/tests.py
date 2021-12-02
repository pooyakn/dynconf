import unittest

from dynconf import Config


class ConfigIntegerTest(unittest.TestCase):
    def setUp(self):
        self.c = Config('/configs/curiosity/')

    def tearDown(self):
        self.c.close()

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
