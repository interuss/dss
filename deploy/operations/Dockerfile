FROM ubuntu:22.04

ENV COCKROACH_VERSION 21.2.7

RUN  apt-get update \
&& apt-get install -y unzip curl gnupg lsb-release apt-transport-https ca-certificates

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

# Kubectl && Helm
RUN curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | tee /usr/share/keyrings/helm.gpg > /dev/null \
&& echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | tee /etc/apt/sources.list.d/helm-stable-debian.list \
&& curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg \
&& echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.28/deb/ /' | tee /etc/apt/sources.list.d/kubernetes.list \
&& apt-get update \
&& apt-get install -y kubectl helm

# Cockroach
RUN curl "https://binaries.cockroachdb.com/cockroach-v${COCKROACH_VERSION}.linux-amd64.tgz" -o "cockroach-v${COCKROACH_VERSION}.tgz" \
&& tar -xvf "cockroach-v${COCKROACH_VERSION}.tgz" \
&& cp -i cockroach-v${COCKROACH_VERSION}.*/cockroach /usr/local/bin/ \
&& mkdir -p /usr/local/lib/cockroach \
&& cp -i cockroach-v${COCKROACH_VERSION}.*/lib/libgeos.so /usr/local/lib/cockroach/ \
&& cp -i cockroach-v${COCKROACH_VERSION}.*/lib/libgeos_c.so /usr/local/lib/cockroach/

# TODO: Migrate scripts to python3 commands
RUN ln -s /usr/bin/python3 /usr/bin/python & \
    ln -s /usr/bin/pip3 /usr/bin/pip

# Clean up apt
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*

RUN terraform --version
