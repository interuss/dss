#!/usr/bin/env python3

import argparse
import logging
import shutil
import sys

from apply import do_apply
from cluster import Cluster
from init import do_init
from nodes import do_generate_nodes
from ca_pool import do_get_pool_ca, do_get_ca, do_add_cas, do_list_pool_ca, do_remove_cas

l = logging.getLogger(__name__)


def parse_args():
    parser = argparse.ArgumentParser(
        description="Manage certificates for a yugabyte cluster"
    )
    parser.add_argument(
        "--name",
        metavar="NAME",
        required=True,
        help="name of your cluster, should be unique to identify it",
    )
    parser.add_argument(
        "--organization",
        metavar="ORGANIZATION",
        default="generic-dss-organization",
        help="name of your origanization",
    )
    parser.add_argument(
        "--cluster-context",
        metavar="CLUSTER_CONTEXT",
        required=True,
        help="kubernetes cluster context name",
    )
    parser.add_argument(
        "--namespace",
        metavar="NAMESPACE",
        required=True,
        help="kubernetes cluster namespace you are deploying to.",
    )
    parser.add_argument(
        "--nodes-count",
        metavar="NODES_COUNT",
        default="3",
        help="Number of yugabyte nodes in the cluster, default to 3",
    )
    parser.add_argument(
        "--nodes-public-address",
        metavar="NODES_PUBLIC_ADDRESS",
        default="",
        help="Public node address. Use <ID> to indicate id of the node (0, 1, ...), <TYPE> for the type (tserver, masters). Example: '<ID>.<TYPE>.db.interuss.example'",
    )
    parser.add_argument(
        "--ca-file",
        metavar="CA_FILE",
        default="-",
        help="CA file, for add/remove operation. Set to '-' to use stdin",
    )
    parser.add_argument(
        "--ca-serial",
        metavar="CA_SERIAL",
        help="CA serial, for remove operation. If set, --ca-file is ignored",
    )
    parser.add_argument(
        "action",
        type=str,
        help="action to be run",
        choices=[
            "init",
            "apply",
            "regenerate-nodes",
            "add-pool-ca",
            "remove-pool-ca",
            "list-pool-ca",
            "get-pool-ca",
            "get-ca",
            "destroy",
        ],
    )
    parser.add_argument(
        "--log-level",
        type=str,
        help="logging level",
        default="INFO",
        choices=[
            "DEBUG",
            "INFO",
            "WARNING",
            "ERROR",
        ],
    )
    return parser.parse_args()


def main():

    args = parse_args()
    logging.basicConfig(
        level=args.log_level,
        format="%(asctime)-15s %(funcName)-25s %(levelname)-8s %(message)s",
    )
    cluster = Cluster(
        args.name,
        args.cluster_context,
        args.namespace,
        args.organization,
        args.nodes_count,
        args.nodes_public_address,
    )

    def read_input():
        if args.ca_file == "-":
            return sys.stdin.read()

        with open(args.ca_file, 'r') as f:
            return f.read()

    if args.action == "init":
        do_init(cluster)
    elif args.action == "regenerate-nodes":
        do_generate_nodes(cluster)
    elif args.action == "apply":
        do_apply(cluster)
    elif args.action == "add-pool-ca":
        do_add_cas(cluster, read_input())
    elif args.action == "remove-pool-ca":
        if args.ca_serial:
            do_remove_cas(cluster, args.ca_serial)
        else:
            do_remove_cas(cluster, read_input())
    elif args.action == "list-pool-ca":
        do_list_pool_ca(cluster)
    elif args.action == "get-pool-ca":
        do_get_pool_ca(cluster)
    elif args.action == "get-ca":
        do_get_ca(cluster)
    elif args.action == "destroy":
        if input("Are you sure? You will loose all your certificates! [yN]") == "y":
            shutil.rmtree(cluster.directory)
            l.warning(f"Destroyed cluster certificates")
        else:
            l.info(f"Cancelled removal")


if __name__ == "__main__":
    main()
