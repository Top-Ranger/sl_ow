FROM golang
COPY . /sl_ow
WORKDIR /sl_ow
RUN go build -o sl_ow
CMD ./sl_ow