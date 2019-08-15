# k8s-node-watcher-example

Sample code that shows how [Notify17](https://notify17.net/)'s [raw API keys](https://notify17.net/docs/raw-api-keys/) can be used to watch a Kubernetes cluster's nodes, when they're added or deleted, and receive a notification on each of these events.

Source code is in the [main.go](./main.go) file, Go version >= 1.12.

To deploy:

1. Create a raw API key in [Notify17's dashboard](https://dash.notify17.net/#/rawAPIKeys).
2. Customize the `k8s-node-watcher-example.yaml` and replace the `REPLACE_API_KEY` environment variable value with your raw API key.
3. Run `kubectl apply -f k8s-node-watcher-example.yaml`

You will then receive a notification for each node that gets added or deleted in the cluster.

When done, you can safely delete the resources by running `kubectl delete -f k8s-node-watcher-example.yaml`.