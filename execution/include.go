package execution

// Include / IncludeMany hydration will follow SqlKata.Execution.handleIncludes:
// after Get, related queries run with WhereIn(foreignKey, localKeys) and results
// are attached by relation name. Typed struct hydration needs reflection + db tags;
// map[string]any hydration is the first target. Not implemented in this initial cut.
