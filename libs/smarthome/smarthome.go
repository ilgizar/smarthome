package smarthome

// Types of a days

var TypesOfDays = []string {
    "leave",            // work vacation
    "vacation",         // school vacation
    "holiday",          // holiday
    "workday",          // work day
}

// Nodes configuration file structure

type NodeOptionConfigStruct struct {
    Name         string
    Value        string
}

type NodeConfigStruct struct {
    Name         string
    Title        string
    IP           string
    Protocol     string
    Option       []NodeOptionConfigStruct
}

type NodesConfigStruct struct {
    Node         []NodeConfigStruct
}

// Usage configuration file structure

type UsageConfigPeriodStruct struct {
    Begin        int
    End          int
}

type UsageConfigConditionStruct struct {
    Date         []string
    DatePeriod   []UsageConfigPeriodStruct
    Time         []string
    TimePeriod   []UsageConfigPeriodStruct
    Weekday      []string
    Weekdays     []string
    Online       []string
    Offline      []string
    Quiet        bool
    Message      string
}

type UsageConfigLimitedStruct struct {
    UsageConfigConditionStruct
    Overall      int
    Using        int
    Pause        int
}

type UsageConfigOnlineStruct struct {
    UsageConfigConditionStruct
    Pause        int
}

type UsageConfigRuleStruct struct {
    Name         string
    Title        string
    Nodes        []string
    Allowed      []UsageConfigConditionStruct
    Denied       []UsageConfigConditionStruct
    Limited      []UsageConfigLimitedStruct
    Online       []UsageConfigOnlineStruct
    Offline      []UsageConfigOnlineStruct
}

type UsageConfigStruct struct {
    Rule         []UsageConfigRuleStruct
}
