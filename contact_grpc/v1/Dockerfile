FROM alpine

RUN apk --update --no-cache add ca-certificates openssl

# Referenced from https://github.com/grpc/grpc/blob/master/tools/dockerfile/test/python_alpine_x64/Dockerfile:
RUN apk --no-cache add \
    autoconf automake bzip2 build-base cmake ccache gcc libtool linux-headers \
    make perl strace unzip wget zip

RUN apk --no-cache add python3 python3-dev py3-pip

WORKDIR /app

COPY requirements/ ./requirements/
COPY requirements.txt .
RUN /usr/bin/pip3 install -r requirements.txt

COPY bobsknobshop/ ./bobsknobshop/
COPY contact_server.py .

CMD ["/usr/bin/python3", "contact_server.py"]
