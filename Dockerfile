FROM golang:bullseye

WORKDIR /usr/src/df2

RUN echo "deb http://deb.debian.org/debian bullseye main contrib non-free" > /etc/apt/sources.list && \
 	echo "deb http://security.debian.org/debian-security/ bullseye-security main contrib non-free" >> /etc/apt/sources.list && \
 	echo "deb http://deb.debian.org/debian bullseye-updates main contrib non-free" >> /etc/apt/sources.list

# file, webp = cwebp cmd?

# apt-utils is required for ansilove deb libs
# libgd-dev and libgd3 are required for the ansilove build
RUN set -x && \
	apt-get autoremove && \
	apt-get update --quiet && \
	apt-get install --quiet --assume-yes apt-utils && \
	apt-get install --quiet --assume-yes \
	file \
	pngquant \
	webp \
	imagemagick \
	unrar \
	unzip \
	arj \
	netpbm \
	lhasa \
	net-tools \
	ansilove && \
	apt-get upgrade --quiet --assume-yes && \
	mkdir -p /root/.cache/df2 && \
	cp -v /usr/bin/cwebp /root/.cache/df2/cwebp

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN git describe --tags --abbrev=0 > .version && \
	go build -race -v -o /usr/local/bin/df2 && \
	/usr/local/bin/df2 config setdb --host=host.docker.internal

ENV DF2_HOST=host.docker.internal

CMD bash