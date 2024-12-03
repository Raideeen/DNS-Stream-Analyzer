import json

import dns.resolver
from flask import Flask, jsonify, request

# TODO : use gRPC for creating the server for handling all the DNS request (raw
# TODO : binary DNS <- more realistic / harder or JSON DNS <- less realistic / easier)
app = Flask(__name__)


@app.route("/health")
def health():
    return "200 OK"


@app.route("/dns-query", methods=["POST"])
def receive_dns_query():
    data = request.json
    # Perform analysis
    print(data)
    return jsonify({"status": "received"}), 200


if __name__ == "__main__":
    app.run(port=8080)
