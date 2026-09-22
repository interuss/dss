# Design of user documentation

## Definitions

The following definitions used:

* A "DSS instance deployment" is a fully-commissioned, working instance of our DSS software, ready to fully meet all applicable product requirements
    * So, for instance, `tk apply` doesn't necessarily achieve a completed DSS deployment if eviction needs to be set up before the product will meet all of its applicable requirements, like maintaining a reasonable level of performance over a long period of time.
    * Since a DSS instance cannot exist without a pool and we do not support changing which pool a DSS instance uses after deployment, at least the initial pooling configuration is a component of deployment.
* "Infrastructure" is all of the cloud resources and configurations necessary to accept deployment of standard services (Kubernetes cluster, static IP addresses, DNS configurations, etc).
* "Services" is all of the executables necessary to produce a complete instance of our product when executed on a suitable infrastructure.
* "Operations" are actions performed on a DSS deployment that don't decommission that deployment.
    * So, actions taken during the process of deploying a DSS instance are not "operations"; they are part of deployment.
* "Decommissioning" actions are performed on a DSS deployment to reduce the scope of a DSS deployment; especially to destroy that deployment entirely.
* "User" in this context is someone who wants to operate a DSS instance, but does not want to modify any part of the product (change software, change deployment tooling, etc).
* "User documentation" is the documentation relevant to the user defined above.
    * Therefore, documentation only relevant to someone wanting to develop the software is not "user documentation".  For instance, user documentation should probably not ask the user to install Go (but developer documentation likely would).
* "Background" is information useful to a user (as defined above) that is not part of procedural documentation.
* "Procedural documentation" is described in the section below.

## User profile

The user for all of the journeys below is assumed to be USS personnel with:
* A moderate level of general technical knowledge (sufficient to design, deploy, and operate the rest of their USS system apart from the DSS)
* A moderate level of familiarity with the relevant standard (e.g., ASTM F3411-22a), since the rest of their USS is implemented according to that standard
* A clear understanding of what the overall system needs to accomplish (e.g., scale and profile of load)
* Little to no knowledge of the InterUSS DSS implementation
    * They will learn the information necessary to use our product from this documentation

## Procedural documentation

All of the user journeys below except background seek to accomplish a particular goal.  To meet this need, documentation should present a clear and complete procedure for the user to follow.  Starting at the top index of the user documentation, a user should be able to easily follow a linear path of actions to accomplish the goal, like a machine interpreter following a compiled binary.  Procedural documentation is source code for human actions.

The user can be easily directed to different documents (and sections of documents, to some extent) like a "goto" machine instruction.

The user can be presented with "if blocks" as long as the user can unambiguously evaluate the condition, easily determine where to go under each condition, and easily see when the "if block" ends and where to go next.

The user can be referred to "subroutines", but they must be provided with all information necessary to complete that "subroutine" ("arguments") before being referred, and the "subroutine" must have a clear endpoint that the user easily understands to mean that they return to where they were before entering the "subroutine".  We should ideally limit the "stack depth" as much as practical, however, as keeping a large stack in working memory can be challenging for humans.

At every step in the entire procedure, the user must have all the information necessary to execute that step before the step is introduced.  If they are missing any information, an additional step must be added prior to that step to acquire the information.

At every step in the entire procedure from the entrypoint to completion of the user journey, the next step the user should take must be clear.

Noting to the user where they can learn more background information when they are interested is useful, but we should minimize the amount of mandatory work required to use our products.  "Go learn how to use this proprietary tool well enough to translate your desired outcomes into tool usages" is highly undesirable; we should instead provide as close to the exact command the user should run using the proprietary tool as practical.

Not all future user documentation needs to be procedural or background, but there should be a clear distinction between procedural and non-procedural documentation.  For instance, "tool reference" describing parameters of command line tools and what they do would be useful non-procedural documentation if users might use those tools outside a procedure defined in procedural documentation, or if there are so many different outcomes different users may want to achieve that understanding the scope and capabilities of a tool is necessary for a user to select their desired outcome.  However, if there isn't a user journey that involves knowing the details of tool usage, full documentation of the tool would likely be better suited to remain solely inside developer documentation.

## User journeys

### Top-level user journeys

The documentation in this folder is built around a few primary user journeys:
* USS wants to deploy a DSS instance
* USS wants to operate an existing DSS instance
* USS wants to decommission an existing DSS instance
* USS wants to know more about the concepts behind parts of the product, why procedures are defined as they are, how they might deviate from InterUSS recommendations to fit their particular use case, etc

These user journeys are fulfilled by the deployment, operations, decommissioning, and background folders.

### Deployment user journey

The deployment journey is split into four modules which are intended to be mostly separable.  For instance, "services" describes how to start executables on a common infrastructure substrate, mostly regardless of how that infrastructure was created.  There are some inter-module dependencies: for instance, service deployment via the tanka files automatically generated by terraform can only be used if terraform generated those files during infrastructure creation.

* Infrastructure is the Kubernetes cluster and other cloud or local resources needed to run the services executables.
* Pooling is the set of steps necessary to configure how the services will establish or join the pool when they are first started.
    * This is where the potentially-blocking async steps of coordinating with existing pool participants are located.
* Services is all the necessary executables, including support services like database eviction and monitoring.
* Verification is the check that the final product produced from all the previous modules was successfully deployed and is suitable for purpose.
    * The configuration or other characteristics of the DSS deployment should not change after verification starts since doing so would mean the verification did not verify the final form of the deployment.
