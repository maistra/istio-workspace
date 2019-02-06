FROM centos:7

RUN cd /tmp \
    && curl -L https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz > openshift-origin-client-tools.tar.gz \
    && tar xvzf openshift-origin*.tar.gz \
    && mv openshift-origin*/oc /usr/bin \
    && rm -rfv openshift-origin*

RUN adduser istio-workspace
USER istio-workspace

ADD dist/ike /usr/local/bin/ike
