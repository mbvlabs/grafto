# grafto


## Views
You can define `partials`, either using `unrolled/render`'s `partial_name-current_tmpl_name` or the one built in with
Go's template library, using `define`. A `define` can be reused throughout the templates by using either `template` or
`block`. Those two are effectively the same, but `block` lets you define a fallback. If you create a file under `partials/`
and put the content inside a `define`, you can use it anywhere by doing `template name`. (TODO look up why) Using 
`unrolled/render`, the `block` override only works when its defined inside a template __not__ in a layout file. I.e.
creating a `block` inside `layouts/base.html` will not be overridable. If you add a `block` to a `define` you can use 
that to add additional elements.
