package smarthome

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
