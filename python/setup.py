from distutils.core import setup

setup(
    name='dynconf',
    version='0.0.1',
    packages=['dynconf'],
    author='Marsel Mavletkulov',
    url='https://github.com/marselester/dynconf',
    description='Dynamic service configuration backed by etcd',
    long_description=open('README.rst').read(),
    install_requires=[
        'etcd3>=0.12.0',
        'python-dateutil>=2.8.2',
        'six',
    ],
    classifiers=[
        'Intended Audience :: Developers',
        'Operating System :: OS Independent',
        'Programming Language :: Python',
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3.5',
        'Programming Language :: Python :: 3.6',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3.8',
        'Topic :: Software Development :: Libraries :: Python Modules'
    ]
)
