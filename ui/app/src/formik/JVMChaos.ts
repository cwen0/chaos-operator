/**
 * This file was auto-generated by @ui/openapi.
 * Do not make direct changes to the file.
 */

export const actions = ['latency', 'return', 'exception', 'stress', 'gc', 'ruleData', 'mysql'],
  data = [
    {
      field: 'text',
      label: 'class',
      value: '',
      helperText: 'Optional.  Java class',
    },
    {
      field: 'label',
      label: 'containerNames',
      value: [],
      helperText:
        'Optional. ContainerNames indicates list of the name of affected container. If not set, the first container will be injected',
    },
    {
      field: 'number',
      label: 'cpuCount',
      value: 0,
      helperText: 'Optional.  the CPU core number needs to use, only set it when action is stress',
    },
    {
      field: 'text',
      label: 'database',
      value: '',
      helperText: 'the match database default value is "", means match all database',
    },
    {
      field: 'text',
      label: 'exception',
      value: '',
      helperText:
        'Optional.  the exception which needs to throw for action `exception` or the exception message needs to throw in action `mysql`',
    },
    {
      field: 'number',
      label: 'latency',
      value: 0,
      helperText:
        "Optional.  the latency duration for action 'latency', unit ms or the latency duration in action `mysql`",
    },
    {
      field: 'text',
      label: 'memType',
      value: '',
      helperText:
        "Optional.  the memory type needs to locate, only set it when action is stress, the value can be 'stack' or 'heap'",
    },
    {
      field: 'text',
      label: 'method',
      value: '',
      helperText: 'Optional.  the method in Java class',
    },
    {
      field: 'text',
      label: 'mysqlConnectorVersion',
      value: '',
      helperText: 'the version of mysql-connector-java, only support 5.X.X(set to "5") and 8.X.X(set to "8") now',
    },
    {
      field: 'text',
      label: 'name',
      value: '',
      helperText: 'Optional.  byteman rule name, should be unique, and will generate one if not set',
    },
    {
      field: 'number',
      label: 'pid',
      value: 0,
      helperText: 'the pid of Java process which needs to attach',
    },
    {
      field: 'number',
      label: 'port',
      value: 0,
      helperText: 'Optional.  the port of agent server, default 9277',
    },
    {
      field: 'text',
      label: 'remoteCluster',
      value: '',
      helperText: 'Optional. RemoteCluster represents the remote cluster where the chaos will be deployed',
    },
    {
      field: 'text',
      label: 'ruleData',
      value: '',
      helperText: "Optional.  the byteman rule's data for action 'ruleData'",
    },
    {
      field: 'text',
      label: 'sqlType',
      value: '',
      helperText:
        "the match sql type default value is \"\", means match all SQL type. The value can be 'select', 'insert', 'update', 'delete', 'replace'.",
    },
    {
      field: 'text',
      label: 'table',
      value: '',
      helperText: 'the match table default value is "", means match all table',
    },
  ]
