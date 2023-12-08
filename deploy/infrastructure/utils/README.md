# Terraform Utils

This directory contains the following tools to simplify the management of the terraform modules:

1. `generate_terraform_variables.sh`: Terraform variables can't be shared between modules without repeating their definition at every level of encapsulation.
   To prevent repeating ourselves and to maintain a consistent level of quality for every module and dependencies, this script takes variables 
   in the `definitions` directory and creates a `variables.tf` file in each modules with the appropriate content.
