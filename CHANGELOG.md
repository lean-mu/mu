Change Log
=========

All notable changes to the project will be documented in this file.

Versioned according to [Semantic Versioning](http://semver.org/).

## Unreleased

Added:

* Server LB - Added the ability to discover runners from a k8s headless service dynamically, introduced a new enviroment variable "FN_RUNNER_K8S_HEADLESS_SERVICE" which points to the name of the headless service to watch.

Changed:

* Took ownership on the packages - sad to see such a great project abandonned.
