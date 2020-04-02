FROM gcr.io/cloud-builders/kubectl

WORKDIR /

COPY _output/bin/kubeclient /usr/local/bin

CMD ["kubeclient"]
