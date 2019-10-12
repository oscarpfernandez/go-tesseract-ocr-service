#
# Tesseract Hackathon Docker Image
#

FROM ubuntu:18.04

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
  python-tornado \
  wget \
  zlib1g-dev \
  git \
  imagemagick \
  ghostscript \
  tesseract-ocr \
  libtesseract-dev \
  tesseract-ocr-eng \
  tesseract-ocr-fra \
  tesseract-ocr-deu \
  tesseract-ocr-eng

RUN wget -qO- https://dl.google.com/go/go1.13.1.linux-amd64.tar.gz | tar xvz -C /usr/local
ENV PATH $PATH:/usr/local/go/bin

# Set GOPATH
ENV GOPATH /go
ENV PATH /go/bin:$PATH

# Set Tesseract Training data location
ENV TESSDATA_PREFIX /usr/share/tesseract-ocr/4.00/tessdata

# Copy code to image
COPY . /go/src/github.com/oscarpfernandez/go-tesseract-ocr-service

WORKDIR /go/src/github.com/oscarpfernandez/go-tesseract-ocr-service

RUN go install -v -a github.com/oscarpfernandez/go-tesseract-ocr-service/cmd/ocr-service/...

CMD /go/bin/ocr-service

EXPOSE 80
