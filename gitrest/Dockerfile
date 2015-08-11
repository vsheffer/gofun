FROM ubuntu:14.04.2
LABEL go-http-auth.git.tag=e59c15b
MAINTAINER vince.sheffer@bhnetwork.com
ADD ./gitrest /usr/bin/gitrest
ADD ./gitconfig /root/.gitconfig
CMD ["/usr/bin/gitrest", "-repo-dir", "/var/local/data/swagger-specs"]
EXPOSE 8080
