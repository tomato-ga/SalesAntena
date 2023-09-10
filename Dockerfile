FROM python:3.11.5-bullseye

WORKDIR /workdir/python
COPY requirements.txt /workdir/python/
COPY . /workdir/

RUN python -m pip install --upgrade pip
RUN apt update && apt upgrade -y
RUN apt install -y vim cron

RUN pip install -r requirements.txt
