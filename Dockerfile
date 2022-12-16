FROM golang:bullseye

WORKDIR /usr/src/df2

RUN echo "deb http://deb.debian.org/debian bullseye main contrib non-free" > /etc/apt/sources.list && \
 	echo "deb http://security.debian.org/debian-security/ bullseye-security main contrib non-free" >> /etc/apt/sources.list && \
 	echo "deb http://deb.debian.org/debian bullseye-updates main contrib non-free" >> /etc/apt/sources.list

RUN set -x && \
	apt-get autoremove && \
	apt-get update --quiet && \
	apt-get install --quiet --assume-yes apt-utils && \
	apt-get install --quiet --assume-yes \
	ansilove \
	arj \
	file \
	imagemagick \
	lhasa \
	libwebp-dev \
	net-tools \
	netpbm \
	pngquant \
	unrar \
	unzip \
	webp && \
	apt-get upgrade --quiet --assume-yes && \
	mkdir -p /root/.cache/df2 && \
	ln -s /usr/bin/cwebp /root/.cache/df2/cwebp

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# to improve performance remove the -race flag from the 'go build' command.
COPY . .
RUN git describe --tags --abbrev=0 > .version && \
	go build -race -o /usr/local/bin/df2 && \
	/usr/local/bin/df2 config setdb --host=host.docker.internal

ENV DF2_HOST=host.docker.internal

# overwrite the defaults of the webp package, https://github.com/nickalie/go-webpbin
# to obtain a valid LIBWEBP_VERSION value see: https://packages.debian.org/bullseye/libwebp-dev
ENV SKIP_DOWNLOAD=true
ENV VENDOR_PATH="/root/.cache/df2/"
ENV LIBWEBP_VERSION=1.2.1

# clean up
RUN apt-get autoremove --yes && \
 	apt-get clean --quiet --yes && \
 	rm -rf /tmp/* /var/tmp/*

CMD ["/bin/bash", "-c", "df2 -v;echo -e 'in docker, to run the unit tests:\ngo test -failfast ./...\ngo test ./pkg/directories/...\ngo test -run ^TestToWebxp$ github.com/Defacto2/df2/pkg/images';/bin/bash"]
