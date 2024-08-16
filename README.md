# ![dsctriage](img/dscexporter-small.png) Discourse Data Exporter
[![dscexporter](https://snapcraft.io/dscexporter/badge.svg)](https://snapcraft.io/dscexporter)

Export specific Discourse data for analysis

The easiest way to install dscexporter is through snap:

```bash
sudo snap install dscexporter
```

Alternatively you can download this repository and build with `go`:

```bash
git clone https://github.com/lvoytek/discourse-data-exporter.git
cd discourse-data-exporter
go build -o ../dscexporter *.go
```

## Usage
To run the exporter with the default settings, open a terminal and enter:

    dscexporter

By default, the application will connect to a Discourse site running on the local machine, prompt which data should be exported, then print the resulting data for all categories and topics in JSON format. The following arguments can be used to change functionality.

### Server
To use a different Discourse server/website, use the `--discourse.site-url` option, along with the desired base URL. For example:

    dscexporter --discourse.site-url https://meta.discourse.org

### Category
If you want to extract data in a single category, then you can specify it with the `--discourse.category` option with a category slug. For example, to get data from the Ubuntu Discourse `Server` category, run:

    dscexporter --discourse.site-url https://discourse.ubuntu.com --discourse.category server

### Topic
To download data from a specific topic, use the `--discourse.topic` option with the topic's ID. This will override any chosen category when specified. To get data from the [Ubuntu Server Reference topic](https://discourse.ubuntu.com/t/ubuntu-server-reference/29949), run:

    dscexporter --discourse.site-url https://discourse.ubuntu.com --discourse.topic 29949

### Continue Collecting Over Time
By default the exporter runs once and exits. If you want to continue collecting data at a set interval, set the `--data.repeat-collect` flag, and specify an interval in minutes with `--data.collection-interval`. To collect data once every 2 hours indefinitely, run:

    dscexporter --data.repeat-collect --data.collection-interval 120

### Data to Export
Each dataset that can be exported has an option to either export or skip. For example, to specify inclusion of user metadata, run:

    dscexporter --export.users

To skip user metadata, run:

    dscexporter --no-export.users

The possible data to export includes:

| Data to Export | Export Option | Skip Option |
| :------------- | :------------ | :---------- |
| User Metadata | `--export.users` | `--no-export.users` |
| User Metadata | `--export.users` | `--no-export.users` |
| Posts/Comments | `--export.posts` | `--no-export.posts` |
| Topic Edits | `--export.edits` | `--no-export.edits` |

### Export Type
The collected data can be exported to MySQL, CSV, and JSON. Specify the export type with `--data.export-type` and `mysql`, `csv`, or `json`. By default, the exporter displays extracted data in JSON format.

### MySQL-Specific Options
When using MySQL mode, the database info can be specified with `--mysql.database-url`, `--mysql.username`, and `--mysql.password`. The database url defaults to `localhost`.

### CSV-Specific Options
When using CSV mode, all files will be written to a directory which can be specified with `--csv.foldername`. By default, It creates a folder called `out/` in the current directory.

> **Note**:
> When using the snap, only directories contained within `$HOME` can be specified.

### Data Download Rate Limiting
If the Discourse server you are gathering data from requires slower API usage, you can specify a delay between calls in seconds with the `--discourse.rate-limit` option. By default this is 1 second.
