/**
 * @license
 * Copyright (c) 2017 The expand.js authors. All rights reserved.
 * This code may only be used under the BSD style license found at https://expandjs.github.io/LICENSE.txt
 * The complete set of authors may be found at https://expandjs.github.io/AUTHORS.txt
 * The complete set of contributors may be found at https://expandjs.github.io/CONTRIBUTORS.txt
 */

// Const
const env = typeof window !== "undefined" ? window : global;

/**
 * A class used to provide Buffer functionality.
 *
 * @class XPBuffer
 * @since 1.0.0
 * @description A class used to provide Buffer functionality
 * @keywords nodejs, expandjs
 * @source https://github.com/expandjs/xp-buffer/blob/master/lib/index.js
 */
module.exports = env.XPBuffer = Buffer;
