# CRI-O Debug Client

![Release](1.0.0)

## Overview

This project provides a lightweight Go-based client designed to communicate directly with the Unix socket of the **CRI-O** (Container Runtime Interface - Open) used in Kubernetes.

The goal is to assist in **debugging and troubleshooting** CRI-O behavior, especially during **early kubelet startup**, when traditional logging and inspection methods might not be available or sufficient.

## Features

- Connect to the CRI-O Unix domain socket
- Send CRI API requests directly
- Inspect runtime state and responses
- Lightweight and dependency-free
- Helpful for diagnosing kubelet startup failures

## Use Cases

- Investigating pod sandbox creation issues
- Verifying communication between kubelet and CRI-O
- Observing runtime responses without requiring full Kubernetes setup

## Requirements

- Go 1.18+
- Access to the CRI-O Unix socket (usually `/var/run/crio/crio.sock`)