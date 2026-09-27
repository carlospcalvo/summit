package main

var errUsage = errors.New(...)

func run(args []string, stdout, stderr io.Writer) error
