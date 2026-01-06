#!/bin/bash

# This script is used to compile the main app binaries
source ./build/build-core.sh

# This script is used to compile the Proxmania plugin
source ./build/build-proxmania.sh

# This script is used to compile the Agentix plugin
source ./build/build-agentix.sh