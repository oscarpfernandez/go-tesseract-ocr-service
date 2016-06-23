#
# Tesseract Hackathon Docker Image
#

FROM ubuntu:15.10

MAINTAINER oscar.pl.fernandez@gmail.com

# Install essential packages needed for compilatiion / execution of Tesseract.
RUN apt-get update && apt-get install -y \
  autoconf \
  automake \
  autotools-dev \
  build-essential \
  checkinstall \
  libjpeg-dev \
  libpng-dev \
  libtiff-dev \
  libtool \
  libicu-dev \
  libpango1.0-0 \
  libpango1.0-dev \
  icu-devtools \
  python \
  python-imaging \
  python-tornado \
  wget \
  zlib1g-dev \
  liblept4 \
  libleptonica-dev \
  git \
  imagemagick \
  ghostscript \
  tesseract-ocr \
  tesseract-ocr-dev \
  tesseract-ocr-eng \
  tesseract-ocr-fra \
  tesseract-ocr-deu \
  tesseract-ocr-eng


# Install Golang 1.6
RUN wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz
ENV PATH $PATH:/usr/local/go/bin

# Set GOPATH
ENV GOPATH /go
ENV PATH /go/bin:$PATH

# Set Tesseract Training data location
ENV TESSDATA_PREFIX /usr/share/tesseract-ocr

# Copy code to image
COPY . /go/src/github.com/oscarpfernandez/go-tesseract-ocr-service

# Generate static resources
# RUN go get github.com/jteeuwen/go-bindata/...
# RUN go get github.com/elazarl/go-bindata-assetfs/...

RUN cd /go/src/github.com/oscarpfernandez/go-tesseract-ocr-service/vendor/github.com/jteeuwen/go-bindata/ && go install ./...
RUN cd /go/src/github.com/oscarpfernandez/go-tesseract-ocr-service/vendor/github.com/elazarl/go-bindata-assetfs/ && go install ./...

RUN cd /go/src/github.com/oscarpfernandez/go-tesseract-ocr-service && go generate

# compile api source
RUN cd /go/src/github.com/oscarpfernandez/go-tesseract-ocr-service && go install ./...

# set entry-point to start api when docker image is ran
ENTRYPOINT /go/bin/go-tesseract-ocr-service

EXPOSE 80
