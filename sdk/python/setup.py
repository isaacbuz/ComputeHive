"""
Setup configuration for ComputeHive Python SDK
"""

from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="computehive",
    version="1.0.0",
    author="ComputeHive Team",
    author_email="sdk@computehive.io",
    description="Python SDK for ComputeHive distributed compute platform",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/computehive/python-sdk",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: Apache Software License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: System :: Distributed Computing",
    ],
    python_requires=">=3.7",
    install_requires=[
        "requests>=2.28.0",
        "dataclasses>=0.6;python_version<'3.7'",
    ],
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "pytest-cov>=4.0.0",
            "pytest-mock>=3.10.0",
            "black>=22.0.0",
            "flake8>=5.0.0",
            "mypy>=0.990",
            "sphinx>=5.0.0",
            "sphinx-rtd-theme>=1.0.0",
        ],
    },
    entry_points={
        "console_scripts": [
            "computehive=computehive.cli:main",
        ],
    },
    project_urls={
        "Bug Reports": "https://github.com/computehive/python-sdk/issues",
        "Documentation": "https://docs.computehive.io/sdk/python",
        "Source": "https://github.com/computehive/python-sdk",
    },
) 