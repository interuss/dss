# Contributing to this repository

Welcome to the DSS repository and thank you for your interest in contributing to it.

In order to maximize the quality of contributions while keeping the time and energy spent by contributors and committers to a minimum, we kindly ask you to adhere to the practices described below.

## General principles
1. Any change to resources in this repository must be handled through a [Pull Request (PR)](https://docs.github.com/en/get-started/quickstart/contributing-to-projects).

1. No PR on the master branch can be merged without being reviewed.

1. The master branch should remain stable at all times. Before a PR is merged into the master branch, it shall pass the tests described in the [Development Guidelines](./DEVELOPMENT.md). Checks are run automatically on every PR.

1. Before your PR can be accepted, you must submit a Contributor License Agreement (CLA). See [here](https://github.com/interuss/tsc/blob/main/CONTRIBUTING.md#contributor-license-agreement-cla) for more details.

1. The size of individual PRs is a key factor with respect to the quality and efficiency of the reviews. When [Large Contributions](#large-contributions) are required, they should be structured as a series of small and focused PRs. Here are some helpful references:
    - [Strategies For Small, Focused Pull Requests, Steve Hicks](https://artsy.github.io/blog/2021/03/09/strategies-for-small-focused-pull-requests/)
    - [The Art of Small Pull Requests, David Wilson](https://essenceofcode.com/2019/10/29/the-art-of-small-pull-requests/)

### Large contributions
Given the critical nature of the content of this repository, contributions require careful review. To ensure that large contributions don't have a negative impact on the quality of the reviews, the following steps help ensure your contribution gets merged progressively, maximizing knowledge sharing between contributors and committers.

#### Initial discussion
Get in touch with an [InterUSS committer](https://github.com/interuss/tsc/blob/main/CONTRIBUTING.md#committers) using the [mailing list](https://github.com/interuss/tsc#mailing-list) or [Slack](https://github.com/interuss/tsc#slack) so they know what to expect and can provide early feedback and guidance.

#### Design document
Following the initial discussion, the next step will be to create a design document that provides a detailed description of the contribution and outlines clearly how it will integrate and interact with the existing components. The expected content of this design document includes a narrative, overall architecture description, sequence diagrams, and test plans. If applicable, based on the advice of an [InterUSS committer](https://github.com/interuss/tsc/blob/main/CONTRIBUTING.md#committers), this material should be submitted as a PR to the repository in a DESIGN.md file or design directory.

#### Work and PRs breakdown
Upon reviewing your design document, an [InterUSS committer](https://github.com/interuss/tsc/blob/main/CONTRIBUTING.md#committers) may require you to break the contribution into smaller packages, allowing small PRs to be reviewed progressively. See this page for an actual [example](https://github.com/interuss/dss/wiki/Updating-SCD-API).
