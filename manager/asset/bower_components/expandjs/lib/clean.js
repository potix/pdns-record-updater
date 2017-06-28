/**
 * @license
 * Copyright (c) 2017 The expand.js authors. All rights reserved.
 * This code may only be used under the BSD style license found at https://expandjs.github.io/LICENSE.txt
 * The complete set of authors may be found at https://expandjs.github.io/AUTHORS.txt
 * The complete set of contributors may be found at https://expandjs.github.io/CONTRIBUTORS.txt
 */

const assertArgument = require('./assertArgument'),
    isString         = require('./isString'),
    isVoid           = require('./isVoid');

/**
 * Returns `string` trimmed and with consequential whitespaces merged into a single one.
 *
 * ```js
 * XP.clean('  abc  ');
 * // => 'abc'
 *
 * XP.clean('  a  b  c  ');
 * // => 'a b c'
 * ```
 *
 * @function clean
 * @since 1.0.0
 * @category string
 * @description Returns `string` trimmed and with consequential whitespaces merged into a single one
 * @source https://github.com/expandjs/expandjs/blog/master/lib/clean.js
 *
 * @param {string} [string = ""] The target string
 * @returns {string} Returns the cleaned string
 */
module.exports = function clean(string) {
    assertArgument(isVoid(string) || isString(string), 1, 'string');
    return string ? string.trim().replace(/[ ]+/g, ' ') : '';
};