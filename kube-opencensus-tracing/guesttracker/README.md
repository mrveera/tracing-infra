# guesttracker

The guestTracker is a microservice written in golang. It's a part of my experimental effort to implement Open Tracing using linkerd

To inject linkerd proxies


kubectl get -n hackerspace deploy -o yaml \
  | linkerd inject - \
  | kubectl apply -f -
