# combi

Combi is a simple package designed to provide a consistent way of defining commands to run both when called statically from a cli call and when launched interactively from a shell.

This package depends heavily on the following two packages:

 - https://github.com/spf13/cobra - Cobra is used to provide a nice nested cli command structure and posix compliant flags
 - https://github.com/abiosoft/ishell - Ishell provides the interactive shell capability


 Combis goal is to provide a link layer between the above two packages, it is also intended to play nicely with the following packages:

- https://github.com/spf13/viper - Viper is a sister package of cobra and provides configuration management which integrates nicely with cobra, the aim is to keep it that way when using combi

- https://github.com/asaskevich/govalidator - Validation package which uses struct tags to define validation rules, we use the same rule structure to determine if fields are required or not when running commands, we also allow for pluggable validation which plays nicely with govalidator
