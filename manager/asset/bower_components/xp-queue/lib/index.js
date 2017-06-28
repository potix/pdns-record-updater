/**
 * @license
 * Copyright (c) 2017 The expand.js authors. All rights reserved.
 * This code may only be used under the BSD style license found at https://expandjs.github.io/LICENSE.txt
 * The complete set of authors may be found at https://expandjs.github.io/AUTHORS.txt
 * The complete set of contributors may be found at https://expandjs.github.io/CONTRIBUTORS.txt
 */

// Const
const env = typeof window !== "undefined" ? window : global,
    XP    = env.XP || require('expandjs');

/*********************************************************************/

/**
 * A class used to provide live queuing functionality.
 *
 * @class XPQueue
 * @since 1.0.0
 * @description A class used to provide live queuing functionality
 * @keywords nodejs, expandjs
 * @source https://github.com/expandjs/xp-queue/blob/master/lib/index.js
 */
module.exports = env.XPQueue = new XP.Class('XPQueue', {

    /**
     * @constructs
     * @param {Object} [options] The queue's options
     *   @param {boolean} [options.autoStart = false] Specifies if the queue should start automatically
     */
    initialize(options) {

        // Setting
        this.channels = {};
        this.current  = {};
        this.running  = false;
        this.options  = options;

        // Starting
        if (this.options.autoStart) { this.start(); }
    },

    /*********************************************************************/

    /**
     * Adds the provided `func` to the queue.
     *
     * A `channel` can be provided as 1st argument.
     *
     * @method add
     * @param {string} [channel]
     * @param {Function} func
     */
    add(channel, func) {

        // Preparing
        if (XP.isFunction(channel)) { func = channel; channel = undefined; }

        // Asserting
        XP.assertArgument(XP.isVoid(channel) || XP.isString(channel), 1, 'string');
        XP.assertArgument(XP.isFunction(func), 2, 'Function');

        // Pushing
        Array.prototype.push.call(this.channels[channel || ''] = this.channels[channel || ''] || [], func);

        // Calling
        this._next(channel);
    },

    /**
     * Removes all the functions from the queue.
     *
     * A `channel` can be provided as 1st argument.
     *
     * @method clear
     * @param {string} [channel]
     */
    clear(channel) {

        // Asserting
        XP.assertArgument(XP.isVoid(channel) || XP.isString(channel), 1, 'string');

        // Splicing
        Array.prototype.splice.call(this.channels[channel || ''] || [], this.current[channel || ''] ? 1 : 0);
    },

    /**
     * Starts invoking the queued functions.
     *
     * @method start
     */
    start() {

        // Setting
        this.running = true;

        // Calling
        Object.keys(this.channels).forEach(channel => this._next(channel));
    },

    /**
     * Stops invoking the queued functions.
     *
     * @method stop
     */
    stop() {

        // Setting
        this.running = false;
    },

    /*********************************************************************/

    /**
     * Invokes the next function.
     *
     * A `channel` can be provided as 1st argument.
     *
     * @method _next
     * @param {string} [channel]
     * @private
     */
    _next: {
        enumerable: false,
        value(channel) {

            // Asserting
            XP.assertArgument(XP.isVoid(channel) || XP.isString(channel), 1, 'string');

            // Preparing
            if (!channel) { channel = ''; }

            // Checking
            if (!this.running || this.current[channel] || !XP.isArray(this.channels[channel], true)) { return; }

            // Setting
            this.current[channel] = this.channels[channel][0];

            // Calling
            this.current[channel](() => {

                // Splicing
                this.channels[channel].splice(0, 1);

                // Setting
                this.current[channel] = null;

                // Calling
                this._next(channel);
            });
        }
    },

    /*********************************************************************/

    /**
     * The queue's channels.
     *
     * @property channels
     * @type Object
     * @readonly
     */
    channels: {
        set(val) { return this.channels || val; },
        validate(val) { return !XP.isObject(val) && 'Object'; }
    },

    /**
     * The map of currently running functions.
     *
     * @property current
     * @type Object
     * @readonly
     */
    current: {
        set(val) { return this.current || val; },
        validate(val) { return !XP.isObject(val) && 'Object'; }
    },

    /**
     * If set to true, the queue is live.
     *
     * @property running
     * @type boolean
     * @readonly
     */
    running: {
        set(val) { return !!val; }
    }
});
