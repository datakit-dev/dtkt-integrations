<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://assets.datakit.cloud/identity/logo-color-on-blue.svg">
    <img alt="DataKit Logo" src="https://assets.datakit.cloud/identity/logo-color.svg" height="30" />
  </picture>
</p>
<p align="center">Open source & cloud-agnostic data activation tools.</p>

---

# DataKit Integrations

DataKit Integrations extend the functionality of DataKit to any external service by implementing Protobuf based interfaces defined in the [DataKit SDK](https://github.com/datakit-dev/dtkt-sdk/tree/main/proto/dtkt).

[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-blue?style=for-the-badge)](https://conventionalcommits.org)

<!-- [![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/datakit-dev/dtkt-integrations/badge)](https://scorecard.dev/viewer/?uri=github.com/datakit-dev/dtkt-integrations) -->

## Documentation

Comprehensive DataKit Integrations documentation is available at https://withdatakit.com/docs/integrations

### Quick Start

To get started with building and contributing integrations, first follow [CONTRIBUTING.md](CONTRIBUTING.md) for initial repository setup. Private integrations can skip this step.

1. **Install the [DataKit CLI](https://withdatakit.com/docs/cli/dtkt/overview)**
   Refer to the CLI documentation or command help for additional details.

2. Create a package in the language of your choice from our [supported languages](#languages) by running:

```shell
dtkt intgr create Example@0.1.0
```

Follow the prompts to configure your integration.

3.  Integrations are developed by running them in the [DataKit CLI](https://github.com/datakit-dev/dtkt-cli) integration dev mode which watches for file changes, automatically rebuilds your integration, and reconnects it to the DataKit Cloud Network (if applicable):

```shell
dtkt intgr dev
```

### Networking (Zero Trust)

Connecting an integration to DataKit Cloud is easy! Simply `dev` or `run` your integration using the [DataKit CLI](https://github.com/datakit-dev/dtkt-cli) with `-n/--network cloud` and a Zero Trust connection to the DataKit Cloud Network is automatically made (no matter where your integration is running).

Refer to the [DataKit Cloud Network documentation](https://withdatakit.com/docs/cloud-network) for more details.

### Languages

The `dtkt-integrations` project provides support for writing integrations in the following languages:

- **Go**: Full native support.
- **Python**: Experimental support coming soon...
- **TypeScript**: Experimental support coming soon...

### Deployment

Deployment of an integration is **cloud agnostic** allowing you to choose the right configuration for your organization's needs and security considerations. Where you deploy an integration is entirely your choice, however, making an integration available to **DataKit Cloud** requires that it be connected to DataKit's **Cloud Network**.

To make an integration **publicly** available on DataKit Cloud, authors must create a Pull Request (PR) adding their integration to this repository, under the `packages/` directory. The DataKit maintainers will be responsible for hosting such integrations and ensuring their compliance with DataKit's security and operational guidelines.

For DataKit managed hosting of **private** integrations or inquiries about the DataKit Cloud Marketplace, please contact DataKit support <support@withdatakit.com>.

There are two infrastructure options for deploying an integration:

- **Private Deployment**: this may be a local machine or within your own cloud (AWS, GCP, etc.).
- **DataKit Cloud**: a deployment of your integration managed by DataKit Cloud.

## Community

You have questions, need support and or just want to talk about DataKit?

Here are ways to get in touch with the DataKit community:

[![Discord](https://img.shields.io/discord/1246163161891999754?style=for-the-badge&logo=discord&logoColor=white&label=Join%20Discord)](https://dtkt.dev/join-discord)
![Discussions](https://img.shields.io/github/discussions/datakit-dev/dtkt-cli?style=for-the-badge&logo=github&logoColor=white)

## Contributing

We welcome contributions! Please see our [CONTRIBUTING.md](CONTRIBUTING.md) file for guidelines on how to get involved.

## Security

We take security seriously! Please review our [SECURITY.md](SECURITY.md) file for details.

## Legal

### License

DataKit Integrations are licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

Copyright © 2025 DataKit, LLC <legal@withdatakit.com>

### Notice

"DataKit" is a trademark of DataKit, LLC. Unauthorized use of the name, logo, or other trademarks is prohibited.

This project depends on DataKit components licensed under the [DataKit Business Source License 1.1](https://withdatakit.com/legal/bsl).

For more details, see the [NOTICE](NOTICE) file.

---

**Developed with ❤️ by DataKit**
