FROM ubuntu:22.04

RUN  apt-get update \
&& apt-get install -y unzip curl gnupg lsb-release

# Terraform CLI
RUN curl -s https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg \
&& echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list \
&& apt-get update \
&& apt-get install -y terraform

# AWS CLI
WORKDIR /opt
RUN  curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" \
&& unzip awscliv2.zip \
&& rm awscliv2.zip \
&& ./aws/install

# Clean up apt
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

RUN terraform --version