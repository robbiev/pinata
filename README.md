Package pinata is a utility to beat data out of `interface{}`, `[]interface{}` and `map[string]interface{}`.

Unlike other packages most methods do not return an error type. They become a no-op when the first error is found so the error can be checked after a series of operations instead of at each operation separately (inspired by https://blog.golang.org/errors-are-values). Special care is taken to return good errors so you can still find out where things went wrong.

See https://godoc.org/github.com/robbiev/pinata for more information.

<a href="https://godoc.org/github.com/robbiev/pinata"><img src="http://garbagecollected.org/img/pinata.jpg" width="40%" height="40%"></a>

(image by [Raquel Gonzalez](https://www.flickr.com/photos/raquelgo/7001286319/))
