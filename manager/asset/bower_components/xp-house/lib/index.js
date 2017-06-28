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
    Room      = require('./room');

/*********************************************************************/

/**
 * A class used to provide rooming functionality.
 *
 * @class XPHouse
 * @extends XPEmitter /bower_components/xp-emitter/lib/index.js
 * @since 1.0.0
 * @description A class used to provide rooming functionality
 * @keywords nodejs, expandjs
 * @source https://github.com/expandjs/xp-house/blob/master/lib/index.js
 */
module.exports = env.XPHouse = new XP.Class('XPHouse', {

    // EXTENDS
    extends: XPEmitter,

    /*********************************************************************/

    /**
     * Emitted when a room has been closed.
     *
     * @event close
     * @param {Object} room
     */

    /**
     * Emitted when a roomer joins a room.
     *
     * @event join
     * @param {Object} room
     * @param {Object} roomer
     */

    /**
     * Emitted when a roomer leaves a room.
     *
     * @event leave
     * @param {Object} room
     * @param {Object} roomer
     */

    /*********************************************************************/

    /**
     * @constructs
     */
    initialize() {

        // Super
        XPEmitter.call(this);

        // Setting
        this.floors = {};

        // Listening
        this.on('close', this._handleClose.bind(this));
    },

    /*********************************************************************/

    /**
     * Checks if a room exists.
     *
     * A 2nd parameter can be provided to specify the `room`'s `floor`.
     *
     * @method isRoom
     * @param {string} name
     * @param {string} [floor]
     * @returns {boolean}
     */
    isRoom(name, floor) {
        return !!this.floors[floor || ''] && !!this.floors[floor || ''][name];
    },

    /**
     * Returns the house's rooms.
     *
     * A 2nd parameter can be provided to specify an house's `floor`.
     *
     * @method rooms
     * @param {string} [floor]
     * @returns {Array}
     */
    rooms(floor) {
        return XP.values(this.floors[floor || ''] || {});
    },

    /*********************************************************************/

    /**
     * Iterates over `rooms`'s roomers, removing the first roomer `identity` returns truthy for.
     * The `callback` is invoked with 2 arguments: (`error`, `success`).
     *
     * A 3rd parameter can be provided to specify the `room`'s `floor`.
     *
     * @method kick
     * @param {Function | Object} identity
     * @param {string} room
     * @param {string} [floor]
     * @param {Function} [callback]
     */
    kick: {
        callback: true,
        value(identity, room, floor, callback) {

            // Asserting
            if (!XP.isString(room, true)) { callback(XP.error(500, '"room" must be string.')); return; }
            if (!XP.isFunction(identity) && !XP.isObject(identity)) { callback(XP.error(500, '"identity" must be Function or Object.')); return; }
            if (!XP.isVoid(floor) && !XP.isString(floor)) { callback(XP.error(500, '"floor" must be string.')); return; }

            // Let
            floor = this.floors[floor || ''] || {};

            // Checking
            if (!floor[room]) { callback(XP.error(404, `Room "${room}" not found.`)); return; }

            // Kicking
            floor[room].kick(identity, callback);
        }
    },

    /**
     * Ensures the requested `room`.
     * The `callback` is invoked with two arguments: (`error`, `room`).
     *
     * A 2nd parameter can be provided to override the `room`'s behavior:
     * - `autoClose`: specifies if the `room` should be closed when the last roomer has been kicked
     * - `cursor`: an object with a `close` method, invoked on the `room`'s closure
     * - `floor`: a namespace for the requested `room`
     *
     * @method room
     * @param {string} name
     * @param {Object} [options]
     *   @param {boolean} [options.autoClose = false]
     *   @param {Object} [options.cursor]
     *   @param {string} [options.floor]
     * @param {Function} [callback]
     */
    room: {
        callback: true,
        value(name, options, callback) {

            // Preparing
            if (XP.isVoid(options)) { options = {}; }

            // Asserting
            if (!XP.isString(name, true)) { callback(XP.error(500, '"name" must be string.')); return; }
            if (!XP.isObject(options)) { callback(XP.error(500, '"options" must be Object.')); return; }
            if (!XP.isVoid(options.cursor) && !XP.isObject(options.cursor)) { callback(XP.error(500, '"options.cursor" must be Object.')); return; }
            if (!XP.isVoid(options.floor) && !XP.isString(options.floor)) { callback(XP.error(500, '"options.floor" must be string.')); return; }

            // Let
            let floor = this.floors[options.floor || ''] = this.floors[options.floor || ''] || {};

            // Opening
            let room = floor[name] || new Room(name, options);

            // Checking
            if (!floor[name]) {

                // Listening
                room.on('close', () => this.emit('close', room));
                room.on('join', roomer => this.emit('join', room, roomer));
                room.on('leave', roomer => this.emit('leave', room, roomer));

                // Setting
                floor[name] = room;
            }

            // Callback
            callback(null, room);
        }
    },

    /*********************************************************************/

    /**
     * The house's floors, used as namespaces for rooms.
     *
     * @property floors
     * @type Object
     * @readonly
     */
    floors: {
        set(val) { return this.floors || val; },
        validate(val) { return !XP.isObject(val) && 'Object'; }
    },

    /*********************************************************************/

    // HANDLER
    _handleClose(room) {

        // Deleting
        delete this.floors[room.floor || ''][room.name];
    }
});
