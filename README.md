# Dstributed ping

[![AI Slop Inside](https://sladge.net/badge.svg)](https://sladge.net)

## Overview

Distributed Ping is a system for performing network reachability checks 
from multiple independent nodes.

Each participant runs a lightweight containerized node that connects to a 
central controller, receives ping tasks, executes them from its own 
network location, and reports the results back. Any connected node can 
initiate checks and request measurements from other nodes in the network.

This makes it possible to verify whether a host is reachable not just 
from a single machine, but from a distributed set of vantage points. 
The system is useful for diagnosing regional connectivity issues, 
comparing reachability across networks, and building a shared, 
self-hosted ping infrastructure.

