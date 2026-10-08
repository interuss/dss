locals {
    # This is the subnet where Kubernetes workload will be running.
    main_subnet_id = var.vpc_id == "" ? aws_subnet.dss[0].id : var.private_subnet_ids[0]
    # This is the subnet where the ELB will be running.
    public_subnet_id = var.vpc_id == "" ? aws_subnet.dss[0].id : var.public_subnet_ids[0]
    # This is the VPC ID that will be used for the Kubernetes cluster.
    vpc_id = var.vpc_id == "" ? aws_vpc.dss[0].id : var.vpc_id
    # This is the CIDR blocks that will be used for the public access.
    public_access_cidrs = var.vpc_id == "" || var.vpc_id == null ? var.use_public_subnets ? ["0.0.0.0/0"] : [] : []
}