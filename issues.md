


## types are not complete

Either generator fails to produce right types or they are missing from schema. Either way a lot of stuff is typed as `any`, which is tricky to reason about.

## claude desktop validates regexp of tool name differently

it wants `^[a-zA-Z0-9_-]{1,64}$`, while inspector is fine with spaces..