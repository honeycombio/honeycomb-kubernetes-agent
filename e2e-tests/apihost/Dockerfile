FROM python
RUN pip install flask zstd
ADD ./server.py /
ENTRYPOINT ["python", "server.py"]
