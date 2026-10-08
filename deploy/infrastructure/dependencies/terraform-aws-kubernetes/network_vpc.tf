resource "aws_vpc" "dss" {
  # Requirements from https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html

  count = var.vpc_id == "" ? 1 : 0

  cidr_block = "10.0.0.0/16"

  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.cluster_name}-vpc"
  }
}

moved {
  from = aws_vpc.dss
  to   = aws_vpc.dss[0]
}

# Create the single internet gateway for the VPC
resource "aws_internet_gateway" "dss" {
  count = var.vpc_id == "" ? 1 : 0

  vpc_id = aws_vpc.dss[0].id
  tags = {
    Name = "${var.cluster_name}"
  }
}

moved {
  from = aws_internet_gateway.dss
  to   = aws_internet_gateway.dss[0]
}


# Pull in the main route table and make it the route table for the kubernetes cluster
# NOTE: For backward compatibility, this always points to the main route table via the association.main filter
data "aws_route_table" "vpc_main" {
  count = var.vpc_id == "" ? 1 : 0

  vpc_id = aws_vpc.dss[0].id

  filter {
    name   = "association.main"
    values = [true]
  }
}

# Retrieves availability zones from region configured in the provisioner
data "aws_availability_zones" "available" {
  state = "available"
}

# Create the two kubernetes subnets
# NOTE: For backward compatibility, this always points to the main route table via the association.main filter
# Uses the two first availability zones of the region
resource "aws_subnet" "dss" {
  count = var.vpc_id == "" ? 2 : 0

  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.dss[0].cidr_block, 8, count.index)
  vpc_id                  = aws_vpc.dss[0].id
  map_public_ip_on_launch = true

  tags = {
    Name                                        = "${var.cluster_name}-subnet-${count.index}"
    "kubernetes.io/role/elb"                    = 1
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  }
}

# If a VPC ID is provided, we need to retrieve the subnets from the existing VPC, and add the tags to the existing subnets
resource "aws_ec2_tag" "kubernetes_cluster_tags" {
  count = var.vpc_id == "" ? 0 : length(var.private_subnet_ids)

  resource_id = var.private_subnet_ids[count.index]
  key         = "kubernetes.io/cluster/${var.cluster_name}"
  value       = "shared"
}
resource "aws_ec2_tag" "kubernetes_role_tags_1" {
  count = var.vpc_id == "" ? 0 : length(var.private_subnet_ids)

  resource_id = var.private_subnet_ids[count.index]
  key         = "kubernetes.io/role/internal-elb"
  value       = 1
}

resource "aws_ec2_tag" "kubernetes_public_cluster_tags" {
  count = var.vpc_id == "" ? 0 : length(var.public_subnet_ids)

  resource_id = var.public_subnet_ids[count.index]
  key         = "kubernetes.io/cluster/${var.cluster_name}"
  value       = "shared"
}

resource "aws_ec2_tag" "kubernetes_role_tags_2" {
  count = var.vpc_id == "" ? 0 : length(var.public_subnet_ids)

  resource_id = var.public_subnet_ids[count.index]
  key         = "kubernetes.io/role/elb"
  value       = 1
}

# Add a route for the internet gateway into the public route table
resource "aws_route" "internet_gateway" {
  count = var.vpc_id == "" ? 1 : 0

  route_table_id         = data.aws_route_table.vpc_main[0].id
  gateway_id             = aws_internet_gateway.dss[0].id
  destination_cidr_block = "0.0.0.0/0"
}

# Add route table associations for the kubernetes subnets
# NOTE: For backward compatibility, this always points to the main route table via the association.main filter
# NOTE2: This will be either public or private subnets depending on the use_public_subnets variable
resource "aws_route_table_association" "subnet" {
  count          = var.vpc_id == "" ? 2 : 0
  route_table_id = data.aws_route_table.vpc_main[0].id
  subnet_id      = aws_subnet.dss[count.index].id
}

resource "aws_security_group" "eks-controlplane" {
  description = "Cluster communication with worker nodes"
  vpc_id      = var.vpc_id == "" ? aws_vpc.dss[0].id : var.vpc_id
}

data "aws_vpc" "existing" {
  count = var.vpc_id == "" ? 0 : 1
  id    = var.vpc_id
}

resource "aws_security_group_rule" "eks-controlplane-ingress" {
  description       = var.use_public_subnets ? "Allow traffic from the internet" : "Allow traffic from the VPC"
  type              = "ingress"
  from_port         = 0
  to_port           = 65535
  protocol          = "tcp"
  cidr_blocks       = var.use_public_subnets ? ["0.0.0.0/0"] : [local.vpc_cidr_block]
  security_group_id = aws_security_group.eks-controlplane.id
}
