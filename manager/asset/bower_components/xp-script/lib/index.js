/**
 * @license
 * Copyright (c) 2017 The expand.js authors. All rights reserved.
 * This code may only be used under the BSD style license found at https://expandjs.github.io/LICENSE.txt
 * The complete set of authors may be found at https://expandjs.github.io/AUTHORS.txt
 * The complete set of contributors may be found at https://expandjs.github.io/CONTRIBUTORS.txt
 */

// Const
const env     = typeof window !== "undefined" ? window : global,
    XP        = env.XP || require('expandjs'),
    XPEmitter = env.XPEmitter || require('xp-emitter'),
    promises  = {};

/*********************************************************************/

/**
 * A class used to provide script injection functionality.
 *
 * @class XPScript
 * @extends XPEmitter /bower_components/xp-emitter/lib/index.js
 * @since 1.0.0
 * @description A class used to provide script injection functionality
 * @keywords nodejs, expandjs
 * @source https://github.com/expandjs/xp-script/blob/master/lib/index.js
 */
module.exports = env.XPScript = new XP.Class('XPScript', {

    // EXTENDS
    extends: XPEmitter,

    /*********************************************************************/

    /**
     * Emitted when the script is loaded.
     *
     * @event load
     * @param {*} data
     */

    /**
     * Emitted when the script state changes.
     *
     * @event state
     * @param {string} state
     */

    /*********************************************************************/

    /**
     * @constructs
     * @param {Object} options The script's options
     *   @param {string} src The script's src
     *   @param {string} [callback] The callback to invoke on load
     * @param {Function} resolver The promise callback
     */
    initialize: {
        promise: true,
        value(options, resolver) {

            // Attempting
            XP.attempt(resolver => {

                // Super
                XPEmitter.call(this);

                // Setting
                this.state    = 'idle';
                this.options  = options;
                this.src      = this.options.src;
                this.callback = this.options.callback || null;

                // Let
                let promised = promises.hasOwnProperty(this.src),
                    older    = promised ? null : document.head.querySelector(`script[src="${this.src}"]`),
                    script   = promised ? null : document.createElement(`script`),
                    promise  = promised ? promises[this.src] : new Promise(resolve => (this.callback ? window : script)[this.callback || 'onload'] = resolve);

                // Resolving
                promises[this.src] = promise.then(data => {

                    // Checking
                    if (this.callback) { delete window[this.callback]; } else { data = null; }

                    // Ensuring
                    this.state = 'loaded';
                    this.data  = data;

                    // Emitting
                    this.emit('load', data);

                    // Resolving
                    resolver(null, data);
                });

                // Checking
                if (promised) { return; }

                // Removing
                if (older) { document.head.removeChild(older); }

                // Appending
                document.head.appendChild(Object.assign(script, {async: true, src: this.src}));

                // Setting
                if (this.state !== 'loaded') { this.state = 'pending'; }

            }, resolver);
        }
    },

    /*********************************************************************/

    /**
     * The callback to invoke on load.
     *
     * @property callback
     * @type string
     */
    callback: {
        set(val) { return XP.isDefined(this.callback) ? this.callback : val; },
        validate(val) { return !XP.isNull(val) && !XP.isString(val, true) && 'string'; }
    },

    /**
     * The loaded data.
     *
     * @property data
     * @type *
     * @readonly
     */
    data: {
        set(val) { return XP.isDefined(this.data) ? this.data : val; }
    },

    /**
     * If set to true, the script has been loaded.
     *
     * @property loaded
     * @type boolean
     * @readonly
     */
    loaded: {
        get() { return this.state === 'loaded'; }
    },

    /**
     * The script's src.
     *
     * @property src
     * @type string
     */
    src: {
        set(val) { return this.src || val; },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    },

    /**
     * The script's state.
     *
     * @property state
     * @type string
     * @readonly
     */
    state: {
        set(val) { return val; },
        then(val) { this.emit('state', val); },
        validate(val) { return !this.states.includes(val) && 'string'; }
    },

    /**
     * The list of possible script's states.
     *
     * @property states
     * @type Array
     * @default ["idle", "loaded", "pending"]
     * @readonly
     */
    states: {
        frozen: true,
        writable: false,
        value: ['idle', 'loaded', 'pending']
    }
});
