// Const
const Uglify = require('uglifyjs-webpack-plugin');

// Exporting
module.exports = {
    entry: './index.js',
    externals: {
        'expandjs': 'XP'
    },
    output: {
        filename: 'xp-queue.js',
        path: `${__dirname}/dist`
    },
    plugins: [
        new Uglify({compress: {warnings: false}, output: {comments: false}})
    ]
};
