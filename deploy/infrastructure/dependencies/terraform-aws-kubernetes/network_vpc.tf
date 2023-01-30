resource "aws_vpc" "dss" {
  # Requirements from https://docs.aws.amazon.com/eks/latest/userguide/network_reqs.html
  cidr_block = "10.0.0.0/16"

  enable_dns_hostnames = true
  enable_dns_support = true

  tags = {
    Name = "${var.cluster_name}-vpc"
  }
}

resource "aws_internet_gateway" "dss" {
  vpc_id = aws_vpc.dss.id
  tags = {
    Name = "${var.cluster_name}"
  }
}

data aws_route_table "vpc_main" {
  vpc_id = aws_vpc.dss.id

  filter {
    name   = "association.main"
    values = [true]
  }
}

# Retrieves availability zones from region configured in the provisioner
data "aws_availability_zones" "available" {
  state = "available"
}

# Uses the two first availability zones of the region
resource "aws_subnet" "dss" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.dss.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.dss.id
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.cluster_name}-subnet-${count.index}"
    "kubernetes.io/role/elb"              = 1
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  }
}

# This is the subnet where Kubernetes workload will be running.
data aws_subnet "main_subnet" {
  id = aws_subnet.dss[0].id
}

resource aws_route "internet_gateway" {
  route_table_id = data.aws_route_table.vpc_main.id
  gateway_id = aws_internet_gateway.dss.id
  destination_cidr_block = "0.0.0.0/0"
}

resource aws_route_table_association "subnet" {
  count          = 2
  route_table_id = data.aws_route_table.vpc_main.id
  subnet_id      = aws_subnet.dss[count.index].id
}

resource aws_security_group "eks-controlplane" {
  description = "Cluster communication with worker nodes"
  vpc_id = aws_vpc.dss.id
}
