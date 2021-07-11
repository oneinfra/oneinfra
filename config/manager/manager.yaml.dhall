let oneinfraVersion = (../../dhall/versions.dhall).oneinfraVersion

in  ''
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: controller-manager
      namespace: system
      labels:
        control-plane: controller-manager
    spec:
      selector:
        matchLabels:
          control-plane: controller-manager
      replicas: 1
      template:
        metadata:
          labels:
            control-plane: controller-manager
        spec:
          containers:
          - command:
            - /oi-manager
            args:
            - --enable-leader-election
            image: oneinfra/oi-manager:${oneinfraVersion}
            name: manager
          terminationGracePeriodSeconds: 10
    ''
