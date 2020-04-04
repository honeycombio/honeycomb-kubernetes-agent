# Super simple mock Honeycomb API server
import collections
import flask
import sys
import json
import zstd
import io

app = flask.Flask(__name__)

events = collections.defaultdict(list)


@app.route("/1/batch/<dataset>", methods=['POST'])
def receive_events(dataset):
    data = flask.request.data
    if flask.request.headers.get("Content-Encoding") == "zstd":
      data = zstd.decompress(data)
    data = json.loads(data)
    events[dataset].extend(data)
    resp = len(data) * [{"status": 202}]
    return flask.jsonify(resp)


@app.route("/")
def return_calls():
    return flask.jsonify(events)


if __name__ == '__main__':
    app.run(host='0.0.0.0')
