# UBI 8 with wireguard-go 1.17

FROM registry.access.redhat.com/ubi8/ubi:8.6

RUN dnf install -q -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm https://www.elrepo.org/elrepo-release-8.el8.elrepo.noarch.rpm && \
    dnf install -q -y iproute wireguard-tools && \
    dnf clean all

ADD wireguard-go /usr/local/bin/wireguard-go
ADD wireguard-setup.sh /usr/local/bin/wireguard-setup.sh

CMD wireguard-setup.sh

# UBI 8 with Kernel module -- failing: - nothing provides kernel(genl_unregister_family) = 0xf9388c43 needed by kmod-wireguard-6:1.0.20220627-2.el8_6.elrepo.x86_64 ...

#FROM registry.access.redhat.com/ubi8/ubi:8.6

#RUN dnf install -q -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm https://www.elrepo.org/elrepo-release-8.el8.elrepo.noarch.rpm && \
#    dnf install -q -y kmod-wireguard wireguard-tools && \
#    dnf clean all

#ADD wireguard-setup.sh /usr/local/bin/wireguard-setup.sh

#CMD wireguard-setup.sh


# UBI 9 with wireguard-go 1.18 -- failing: no wireguard-tools package for 9

#FROM registry.access.redhat.com/ubi9/ubi:9.0.0

#RUN dnf install -q -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm https://www.elrepo.org/elrepo-release-9.el9.elrepo.noarch.rpm && \
#    dnf install -q -y iproute wireguard-tools && \
#    dnf clean all

#ADD wireguard-go /usr/local/bin/wireguard-go
#ADD wireguard-setup.sh /usr/local/bin/wireguard-setup.sh

#CMD wireguard-setup.sh