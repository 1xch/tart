## Changelog

### tart 0.0.1 11.02.2020
- refactor & cleanup
- refined  & consolidated api
- eliminated Units & Replacement (for now) 
- added a limited dsl for specifying and parsing time directives (e.g. " modifier ! time ") 
- Get & Set notation for getting and setting time according to directive 
- additional get relations by key: GetRelation
- additional set relations by key: SetRelation(Relation), SetBatch(map[string]Relation),
  SetDirect(time.Time), SetParsedDate(string datetime), SetFloat(unix time)  
- Duration parsing by directive
- subsequent testing changes

### tart 0.0.0 01.07.2019  
- project init
