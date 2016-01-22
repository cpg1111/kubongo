FROM busybox:1.24.1-glibc
MAINTAINER Christian Grabowski https://www.github.com/cpg1111
ADD ./kubongo /opt/bin/kubongo
ENTRYPOINT ["/opt/bin/kubongo"]
