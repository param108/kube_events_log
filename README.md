# Event Reader
reads events from k8smaster and dumps them to file

# Build

make build // builds for amd64 linux arch

go build // builds for local golang

# Running

./eventreader --master <master url> --logfile <log filename>

# Quitting

Either SIGTERM or SIGQUIT will cause the eventreader to gracefully exit.

# TODO

1. Allow access to K8sMaster with proper kubeconfig.
        presently we support access to publicly available cluster.
2. Need to dedup events.
        presently we see that events are printed multiple times.
