FROM python:3.7-alpine

COPY --from=golang:1.14-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

ADD ./requirements.txt /app/requirements.txt
RUN pip install -r /app/requirements.txt
RUN rm -rf __pycache__
ADD . /app
WORKDIR /app

ENTRYPOINT ["python", "generate.py"]
