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
    XPEmitter = env.XPEmitter || require('xp-emitter');

/*********************************************************************/

/**
 * A class used to provide room functionality.
 *
 * @class Room
 * @extends XPEmitter /bower_components/xp-emitter/lib/index.js
 * @since 1.0.0
 * @description A class used to provide room functionality
 * @keywords nodejs, expandjs
 * @source https://github.com/expandjs/xp-house/blob/master/lib/room.js
 */
module.exports = new XP.Class('Room', {

    // EXTENDS
    extends: XPEmitter,

    /*********************************************************************/

    /**
     * Emitted when the room has been closed.
     *
     * @event close
     */

    /**
     * Emitted when a roomer joins the room.
     *
     * @event join
     * @param {Object} roomer
     */

    /**
     * Emitted when a roomer leaves the room.
     *
     * @event leave
     * @param {Object} roomer
     */

    /*********************************************************************/

    /**
     * @constructs
     * @param {string} name The room's name
     * @param {Object} [options] The room's options
     *   @param {boolean} [options.autoClose = false] Specifies if the room should be closed when there are no more `roomers` left
     *   @param {Object} [options.cursor] The cursor assigned to the room
     *   @param {string} [options.floor] The room's floor, used as namespace
     */
    initialize(name, options) {

        // Super
        XPEmitter.call(this);

        // Setting
        this.closed    = false;
        this.roomers   = [];
        this.name      = name;
        this.options   = options;
        this.autoClose = this.options.autoClose || false;
        this.cursor    = this.options.cursor || null;
        this.floor     = this.options.floor || null;
    },

    /*********************************************************************/

    /**
     * Closes the room, kicking each roomer.
     *
     * If the room has a `cursor` with a `close` method, it will be invoked too.
     *
     * The `callback` is invoked with two arguments: (`error`, `success`).
     *
     * @method close
     * @param {Function} [callback]
     */
    close: {
        callback: true,
        value(callback) {

            // Checking
            if (this.closed) { callback(null, false); return; }

            // Setting
            this.closed = true;

            // Kicking
            this.roomers.forEach(roomer => this.kick(roomer));

            // Closing
            if (this.cursor) { this.cursor.close(); }

            // Callback
            callback(null, true);
        }
    },

    /**
     * Iterates over the room's `roomers`, kicking the first one `identity` returns truthy for.
     *
     * The `callback` is invoked with two arguments: (`error`, `success`).
     *
     * @method kick
     * @param {Function | Object} identity
     * @param {Function} [callback]
     */
    kick: {
        callback: true,
        value(identity, callback) {

            // Asserting
            if (!XP.isFunction(identity) && !XP.isObject(identity)) { callback(XP.error(500, '"identity" must be Function or Object.')); return; }

            // Let
            let index = this.roomers.findIndex(XP.isFunction(identity) ? identity : roomer => roomer === identity);

            // Checking
            if (index < 0) { callback(null, false); return; }

            // Splicing
            this.emit('leave', this.roomers.splice(index, 1)[0]);

            // Closing
            if (this.autoClose && this.empty) { this.close(); }

            // Callback
            callback(null, true);
        }
    },

    /**
     * Lets in the provided `roomer`.
     *
     * The `callback` is invoked with two arguments: (`error`, `success`).
     *
     * @method let
     * @param {Object} roomer
     * @param {Function} [callback]
     */
    let: {
        callback: true,
        value(roomer, callback) {

            // Asserting
            if (!XP.isObject(roomer)) { callback(XP.error(500, '"roomer" must be Object.')); return; }

            // Checking
            if (this.closed) { callback(null, false); return; }

            // Let
            let index = this.roomers.indexOf(roomer);

            // Pushing
            if (index < 0) { index = this.roomers.push(roomer) - 1; }

            // Emitting
            this.emit('join', this.roomers[index]);

            // Callback
            callback(null, true);
        }
    },

    /*********************************************************************/

    /**
     * If set to true, the room will be closed when there are no more `roomers` left.
     *
     * @property autoClose
     * @type boolean
     * @default false
     */
    autoClose: {
        set(val) { return XP.isDefined(this.autoClose) ? this.autoClose : !!val; }
    },

    /**
     * The room's roomers.
     *
     * @property roomers
     * @type Array
     * @readonly
     */
    roomers: {
        set(val) { return this.roomers || val; },
        validate(val) { return !XP.isArray(val) && 'Array'; }
    },

    /**
     * If set to true, the room has been closed.
     *
     * @property closed
     * @type boolean
     * @default false
     */
    closed: {
        set(val) { return this.closed || !!val; },
        then(val) { return val && this.emit('close'); }
    },

    /**
     * The cursor assigned to the room.
     *
     * For example: it could be a stream cursor, used by the roomers to listen for updates.
     *
     * @property cursor
     * @type Object
     */
    cursor: {
        set(val) { return XP.isDefined(this.cursor) ? this.cursor : val; },
        validate(val) { return !XP.isNull(val) && !XP.isObject(val) && 'Object'; }
    },

    /**
     * If set to true, there are no `roomers`.
     *
     * @property empty
     * @type boolean
     * @readonly
     */
    empty: {
        get() { return !this.roomers.length; }
    },

    /**
     * The room's floor, used as namespace.
     *
     * @property floor
     * @type string
     * @readonly
     */
    floor: {
        set(val) { return XP.isDefined(this.floor) ? this.floor : val; },
        validate(val) { return !XP.isNull(val) && !XP.isString(val, true) && 'string'; }
    },

    /**
     * The room's name.
     *
     * @property name
     * @type string
     */
    name: {
        set(val) { return this.name || val; },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    }
});
