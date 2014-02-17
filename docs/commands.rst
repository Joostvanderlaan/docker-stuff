.. _commands:

========
Commands
========

The Blockade CLI is built to make it easy to manually manage your containers,
and is also easy to wrap in scripts as needed. All commands that produce
output support a ``--json`` flag to output in JSON instead of plain text.

For the most up to date and detailed command help, use the built-in CLI help
system (``blockade --help``).

``up``
------

::

    usage: blockade up [--json]

    Start the containers and link them together

      --json      Output in JSON format

``destroy``
-----------

::

    usage: blockade destroy

    Destroy all containers and restore networks

``status``
----------

::

    usage: blockade status [--json]

    Print status of containers and networks

    optional arguments:
      --json      Output in JSON format

``logs``
--------

::

    usage: blockade logs CONTAINER

    Fetch the logs of a container

      CONTAINER    Container to fetch logs for

``flaky``
---------

::

    usage: blockade flaky [--all] [CONTAINER [CONTAINER ...]]

    Make the network flaky for some or all containers

      CONTAINER   Container to select

      --all       Select all containers

``slow``
--------

::

    usage: blockade slow [--all] [CONTAINER [CONTAINER ...]]

    Make the network slow for some or all containers

      CONTAINER   Container to select

      --all       Select all containers

``fast``
--------

::

    usage: blockade fast [--all] [CONTAINER [CONTAINER ...]]

    Restore network speed and reliability for some or all containers

      CONTAINER   Container to select

      --all       Select all containers


``partition``
-------------

::

    usage: blockade partition PARTITION [PARTITION ...]

    Partition the network between containers

        Replaces any existing partitions outright. Any containers NOT specified
        in arguments will be globbed into a single implicit partition. For
        example if you have three containers: c1, c2, and c3 and you run:

            blockade partition c1

        The result will be a partition with just c1 and another partition with
        c2 and c3.


      PARTITION   Comma-separated partition

``join``
--------

::

    usage: blockade join

    Restore full networking between containers