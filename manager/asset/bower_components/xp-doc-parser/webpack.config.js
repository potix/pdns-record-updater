// Const
const Uglify = require('uglifyjs-webpack-plugin');

// Exporting
module.exports = {
    entry: './index.js',
    externals: {
        'expandjs': 'XP',
        'xp-emitter': 'XPEmitter',
        'xp-request': 'XPRequest'
    },
    output: {
        filename: 'xp-doc-parser.js',
        path: `${__dirname}/dist`
    },
    plugins: [
        new Uglify({compress: {warnings: false}, output: {comments: false}})
    ]
};
