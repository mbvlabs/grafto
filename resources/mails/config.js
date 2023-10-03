/** @type {import('@maizzle/framework').Config} */

/*
|-------------------------------------------------------------------------------
| Development config                      https://maizzle.com/docs/environments
|-------------------------------------------------------------------------------
|
| The exported object contains the default Maizzle settings for development.
| This is used when you run `maizzle build` or `maizzle serve` and it has
| the fastest build time, since most transformations are disabled.
|
*/

module.exports = {
  locals: {
    site: {
      url: process.env.SCHEME + process.env.HOST,
      name: process.env.PROJECT_NAME,
    }
  },
  build: {
    templates: {
      source: 'src/templates',
      destination: {
        path: '../../pkg/mail/templates/',
      },
      assets: {
        source: 'src/images',
        destination: '../../../static/images',
      },
      plaintext: true,
    },
  },
}
