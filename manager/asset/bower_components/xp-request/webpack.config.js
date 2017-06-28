// Const
const Uglify = require('uglifyjs-webpack-plugin');

// Exporting
module.exports = {
    entry: './index.js',
    externals: {
        'expandjs': 'XP',
        'http': 'http',
        'https': 'https',
        'xp-buffer': 'XPBuffer',
        'xp-emitter': 'XPEmitter'
    },
    output: {
        filename: 'xp-request.js',
        path: `${__dirname}/dist`
    },
    plugins: [
        new Uglify({compress: {warnings: false}, output: {comments: false}})
    ]
};
