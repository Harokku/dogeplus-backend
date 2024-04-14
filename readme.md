# Server Setup Guide

This guide will help you to set up the server step by step.

## Prerequisites

Before you proceed, make sure you have Go SDK version 1.22.2 installed. Refer to the
online [Go documentation](https://golang.org/doc/install/) if you need help with this step.

Additionally, you will need permission to create and work with environment variables on your system.

## Environment Setup

This section describes the environment variables that need to be set up for the server to function correctly.

Variable can be either come from a `config.toml` file or from environment

# Sample TOML Configuration File

```
[variables]
PORT = 3000
SHAREPATH = "/path/to/share"
```