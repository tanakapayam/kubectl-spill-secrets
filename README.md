## NAME
kubectl-spill-secrets - A helper function to see the base64-decoded kubectl secrets

## SYNOPSIS
    kubectl-spill-secrets [-h|--help] [--ejson-public-key=EJSON_PUBLIC_KEY] \
        [--hyphen-to-underscore-keys] [--redacted] [--uppercase-keys]

## DESCRIPTION
This program takes piped stdin from `kubectl get secret SECRET -o yaml`, and
base64-decodes the values in `data` map to stdout. If it is also given a valid
EJSON public key, the output is a plaintext EJSON object -- ready for EJSON
encryption.

## EXAMPLES
    $ kubectl-spill-secrets -h
    $ kubectl get --namespace=APP_NAMESPACE secret APP_SECRET_NAME -o yaml \
       | kubectl-spill-secrets
    $ kubectl get --namespace=APP_NAMESPACE secret APP_SECRET_NAME -o yaml \
       | kubectl-spill-secrets --ejson-public-key=1123498701234098712340985689723641092837415189213419898761234567

## AUTHOR
* Payam Tanaka, @tanakapayam

## REFERENCES
* [https://github.com/Shopify/ejson](https://github.com/Shopify/ejson)
