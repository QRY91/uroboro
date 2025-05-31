#!/usr/bin/env python3
from setuptools import setup, find_packages

setup(
    name="uroboro",
    version="0.1.0",
    description="The Self-Documenting Content Pipeline",
    author="Q",
    author_email="hello@uroboro.dev",
    url="https://github.com/qry91/uroboro",
    packages=find_packages(),
    entry_points={
        'console_scripts': [
            'uroboro=src.cli:main',
            'uro=src.cli:main',  # Short alias for quick usage
        ],
    },
    install_requires=[
        "requests",
        "python-dateutil",
        # Add other dependencies as needed
    ],
    python_requires=">=3.8",
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
    ],
) 