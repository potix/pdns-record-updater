/**
 * @license
 * Copyright (c) 2017 The expand.js authors. All rights reserved.
 * This code may only be used under the BSD style license found at https://expandjs.github.io/LICENSE.txt
 * The complete set of authors may be found at https://expandjs.github.io/AUTHORS.txt
 * The complete set of contributors may be found at https://expandjs.github.io/CONTRIBUTORS.txt
 */

// Const
const env     = typeof window !== "undefined" ? window : global,
    http      = typeof window !== "undefined" ? null : require('http'),
    https     = typeof window !== "undefined" ? null : require('https'),
    codes     = typeof window !== "undefined" ? {} : http.STATUS_CODES,
    location  = env.location || {},
    XP        = env.XP || require('expandjs'),
    XPBuffer  = env.XPBuffer || require('xp-buffer'),
    XPEmitter = env.XPEmitter || require('xp-emitter');

/*********************************************************************/

/**
 * A class used to provide XHR functionality.
 *
 * @class XPRequest
 * @extends XPEmitter /bower_components/xp-emitter/lib/index.js
 * @since 1.0.0
 * @description A class used to provide XHR functionality
 * @keywords nodejs, expandjs
 * @source https://github.com/expandjs/xp-request/blob/master/lib/index.js
 */
module.exports = env.XPRequest = new XP.Class('XPRequest', {

    // EXTENDS
    extends: XPEmitter,

    /*********************************************************************/

    /**
     * Emitted when the entire data is downloaded.
     *
     * @event data
     * @param {*} data
     */

    /**
     * Emitted when a chunk of data is downloaded.
     *
     * @event download
     * @param {number} loaded
     * @param {number} total
     */

    /**
     * Emitted when an error is received.
     *
     * @event error
     * @param {Object} error
     */

    /**
     * Emitted when a response is received.
     *
     * @event response
     * @param {Object} response
     */

    /**
     * Emitted when the request's state changes.
     *
     * @event state
     * @param {string} state
     */

    /**
     * Emitted when the request is submitted.
     *
     * @event submit
     * @param {Buffer | string} data
     */

    /*********************************************************************/

    /**
     * @constructs
     * @param {Object | string} options The request's url or options
     *   @param {Object} [options.headers] An object containing request headers
     *   @param {string} [options.hostname] The request's hostname, usable in alternative to url
     *   @param {number} [options.keepAlive = 0] How often to submit TCP KeepAlive packets over sockets being kept alive
     *   @param {string} [options.method = "GET"] A string specifying the HTTP request method
     *   @param {string} [options.parser = "json"] The type of data expected back from the server
     *   @param {string} [options.path] The request's path, usable in alternative to url
     *   @param {number} [options.port] The request's port, usable in alternative to url
     *   @param {number} [options.protocol = "http:"] The request's protocol, usable in alternative to url
     *   @param {string} [options.url] The request's url
     * @param {Function} resolver The promise callback
     */
    initialize: {
        promise: true,
        value(options, resolver) {

            // Attempting
            XP.attempt(resolver => {

                // Super
                XPEmitter.call(this);

                // Overriding
                if (!XP.isObject(options)) { options = {url: options}; }
                if (!XP.isFalsy(options.url)) { Object.assign(options, XP.pick(XP.parseURL(options.url), ['hostname', 'path', 'port', 'protocol'])); }

                // Setting
                this.downLoaded = 0;
                this.chunks     = [];
                this.state      = 'idle';
                this.options    = options;
                this.headers    = this.options.headers || {};
                this.hostname   = this.options.hostname || location.hostname || '';
                this.keepAlive  = this.options.keepAlive || 0;
                this.method     = this.options.method || 'GET';
                this.parser     = this.options.parser || 'json';
                this.path       = this.options.path || '';
                this.port       = this.options.port || (!this.options.hostname && location.port) || null;
                this.protocol   = this.options.protocol || (!this.options.hostname && location.protocol) || 'http:';
                this.url        = XP.toURL({hostname: this.hostname, pathname: this.pathname, port: this.port, protocol: this.protocol, search: this.search});

                // Binding
                this._handleDownload = this._handleDownload.bind(this);
                this._handleEnd      = this._handleEnd.bind(this, resolver);
                this._handleError    = this._handleError.bind(this, resolver);
                this._handleResponse = this._handleResponse.bind(this, resolver);

            }, resolver);
        }
    },

    /*********************************************************************/

    /**
     * Get or set a header.
     *
     * @method header
     * @param {string} name
     * @param {number | string} [value]
     * @returns {number | string}
     */
    header(name, value) {

        // Asserting
        XP.assertArgument(XP.isString(name, true), 1, 'string');
        XP.assertArgument(XP.isVoid(value) || XP.isFalse(value) || XP.isInput(value, true), 2, 'string');

        // Getting
        if (!XP.isDefined(value)) { return this.headers[name]; }

        // Checking
        if (this.state !== 'idle') { return; }

        // Setting
        if (value) { this.headers[name] = value; } else { delete this.headers[name]; }
    },

    /*********************************************************************/

    /**
     * Aborts the request.
     *
     * @method abort
     * @param {Function} [callback]
     * @returns {Object}
     */
    abort: {
        callback: true,
        value(callback) {

            // Checking
            if (this.tsAbort || !this.tsSubmit) { callback(null, false); return; }

            // Aborting
            this.adaptee.abort();

            // Setting
            this.state   = 'aborted';
            this.tsAbort = Date.now();

            // Callback
            callback(null, true);

            // Chaining
            return this;
        }
    },

    /**
     * Submits the request, using `data` for the request's body.
     *
     * @method submit
     * @param {*} [data]
     * @param {Function} [callback]
     * @returns {Object}
     */
    submit: {
        callback: true,
        value(data, callback) {

            // Checking
            if (!XP.isVoid(data) && !XPBuffer.isBuffer(data) && !XP.isInput(data) && !XP.isCollection(data)) { callback(XP.error(400)); return; }

            // Preventing
            if (this.tsSubmit) { return; }

            // Callback
            this.ready(callback);

            // Let
            let write = ['PATCH', 'POST', 'PUT'].includes(this.method),
                query = write ? '' : XP.toQueryString(data, true),
                body  = write ? data : undefined;

            // Casting
            if (body && XP.isCollection(data) && !XPBuffer.isBuffer(data)) { body = JSON.stringify(data); }

            // CASE: server
            if (http) {

                // Adapting
                let request = this.adaptee = (this.protocol.startsWith('https') ? https : http).request({
                    hostname: this.hostname,
                    keepAlive: this.keepAlive > 0,
                    keepAliveMsecs: this.keepAlive,
                    method: this.method,
                    path: this.pathname + (query || this.search),
                    port: this.port,
                    protocol: this.protocol,
                    withCredentials: false
                });

                // Headers
                Object.keys(this.headers).forEach(key => XP.isInput(this.headers[key], true) && request.setHeader(key, this.headers[key]));

                // Listening
                request.on('error',    this._handleError);
                request.on('response', this._handleResponse);
            }

            // CASE: browser
            if (!http) {

                // Adapting
                let xhr = this.adaptee = new XMLHttpRequest();

                // Opening
                xhr.open(this.method, query ? `${this.url.replace(/\?.*/, ``)}${query}` : this.url, true);

                // Headers
                Object.keys(this.headers).forEach(key => XP.isInput(this.headers[key], true) && xhr.setRequestHeader(key, this.headers[key]));

                // Listening
                xhr.addEventListener('progress', this._handleResponse);
                xhr.addEventListener('progress', this._handleDownload);
                xhr.addEventListener('load', this._handleResponse);
                xhr.addEventListener('load', this._handleEnd);
                xhr.addEventListener('error', this._handleError);
                xhr.upload.addEventListener('error', this._handleError);
            }

            // Setting
            this.state    = 'pending';
            this.tsSubmit = Date.now();

            // Sending
            this.adaptee[http ? 'end' : 'send'](body);

            // Emitting
            this.emit('submit', data);

            // Chaining
            return this;
        }
    },

    /*********************************************************************/

    /**
     * The request's instance.
     *
     * @property adaptee
     * @type Object
     */
    adaptee: {
        set(val) { return this.adaptee || val; },
        validate(val) { return !XP.isObject(val) && 'Object'; }
    },

    /**
     * The received chunks.
     *
     * @property chunks
     * @type Array
     */
    chunks: {
        set(val) { return this.chunks || val; },
        validate(val) { return !XP.isArray(val) && 'Array'; }
    },

    /**
     * The received data.
     *
     * @property data
     * @type *
     * @readonly
     */
    data: {
        set(val) { return XP.isDefined(this.data) ? this.data : val; }
    },

    /**
     * The downloaded bytes.
     *
     * @property downLoaded
     * @type Number
     * @readonly
     */
    downLoaded: {
        set(val) { return val; },
        validate(val) { return !XP.isInt(val, true) && 'number'; }
    },

    /**
     * The download ratio.
     *
     * @property downRatio
     * @type Number
     * @readonly
     */
    downRatio: {
        get() { return XP.isVoid(this.downTotal) ? null : (this.downTotal ? this.downLoaded / this.downTotal : 1); }
    },

    /**
     * The download total.
     *
     * @property downTotal
     * @type Number
     * @readonly
     */
    downTotal: {
        set(val) { return XP.isDefined(this.downTotal) ? this.downTotal : val; },
        validate(val) { return !XP.isNull(val) && !XP.isInt(val, true) && 'number'; }
    },

    /**
     * The received error message.
     *
     * @property error
     * @type Object
     * @readonly
     */
    error: {
        set(val) { return XP.isDefined(this.error) ? this.error : val; },
        validate(val) { return !XP.isNull(val) && !XP.isObject(val) && 'Object'; }
    },

    /**
     * An object containing request headers.
     *
     * @property headers
     * @type Object
     */
    headers: {
        set(val) { return this.headers || (XP.isObject(val) && XP.cloneDeep(val)); },
        validate(val) { return !XP.isObject(val) && 'Object'; }
    },

    /**
     * The request's host.
     *
     * @property host
     * @type string
     * @readonly
     */
    host: {
        get() { return `${this.hostname || ``}${this.port ? `:${this.port}` : ``}`; }
    },

    /**
     * The request's hostname.
     *
     * @property hostname
     * @type string
     */
    hostname: {
        set(val) { return this.hostname || val; },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    },

    /**
     * How often to submit TCP KeepAlive packets over sockets being kept alive.
     *
     * @property keepAlive
     * @type number
     * @default 0
     */
    keepAlive: {
        set(val) { return XP.isDefined(this.keepAlive) ? this.keepAlive : val; },
        validate(val) { return !XP.isInt(val, true) && 'number'; }
    },

    /**
     * A string specifying the HTTP request method.
     *
     * @property method
     * @type string
     * @default "GET"
     */
    method: {
        set(val) { return this.method || XP.upperCase(val); },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    },

    /**
     * The type of data expected back from the server.
     *
     * @property parser
     * @type string
     * @default "json"
     */
    parser: {
        set(val) { return this.parser || val; },
        validate(val) { return !this.parsers.includes(val) && 'string'; }
    },

    /**
     * The list of possible data types to expect back from the server.
     *
     * @property parsers
     * @type Array
     * @default ["buffer", "json", "text"]
     * @readonly
     */
    parsers: {
        frozen: true,
        writable: false,
        value: ['buffer', 'json', 'text']
    },

    /**
     * The request's path.
     *
     * @property path
     * @type string
     */
    path: {
        set(val) { return XP.isDefined(this.path) ? this.path : val; },
        then(val) { let parts = val.match(/([^?]*)(.*)/); this.pathname = parts[1]; this.query = parts[2].slice(1); },
        validate(val) { return !XP.isString(val) && 'string'; }
    },

    /**
     * The request's pathname.
     *
     * @property pathname
     * @type string
     */
    pathname: {
        set(val) { return XP.isDefined(this.pathname) ? this.pathname : val; },
        validate(val) { return !XP.isString(val) && 'string'; }
    },

    /**
     * The request's port.
     *
     * @property port
     * @type number
     */
    port: {
        set(val) { return XP.isDefined(this.port) ? this.port : val; },
        validate(val) { return !XP.isNull(val) && !XP.isNumeric(val, true) && 'number'; }
    },

    /**
     * The request's protocol.
     *
     * @property protocol
     * @type string
     */
    protocol: {
        set(val) { return this.protocol || val; },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    },

    /**
     * The request's query.
     *
     * @property query
     * @type string
     * @readonly
     */
    query: {
        set(val) { return XP.isDefined(this.query) ? this.query : val; },
        validate(val) { return !XP.isString(val) && 'string'; }
    },

    /**
     * The request's response.
     *
     * @property response
     * @type Object
     * @readonly
     */
    response: {
        set(val) { return this.response || val; },
        validate(val) { return !XP.isObject(val) && 'Object'; }
    },

    /**
     * The request's search.
     *
     * @property search
     * @type string
     * @readonly
     */
    search: {
        get() { return this.query ? `?${this.query}` : ``; }
    },

    /**
     * The request's state.
     *
     * @property state
     * @type string
     * @default "idle"
     * @readonly
     */
    state: {
        set(val) { return val; },
        then(val) { return val !== 'idle' && this.emit('state', val); },
        validate(val) { return !this.states.includes(val) && 'string'; }
    },

    /**
     * The list of possible request's states.
     *
     * @property states
     * @type Array
     * @default ["aborted", "complete", "failed", "idle", "pending"]
     * @readonly
     */
    states: {
        frozen: true,
        writable: false,
        value: ['aborted', 'complete', 'failed', 'idle', 'pending']
    },

    /**
     * The request's status code.
     *
     * @property statusCode
     * @type number
     * @readonly
     */
    statusCode: {
        set(val) { return this.statusCode || val; },
        validate(val) { return !XP.isInt(val, true) && 'number'; }
    },

    /**
     * The request's status code.
     *
     * @property statusMessage
     * @type string
     * @readonly
     */
    statusMessage: {
        set(val) { return XP.isDefined(this.statusMessage) ? this.statusMessage : val; },
        validate(val) { return !XP.isNull(val) && !XP.isString(val) && 'string'; }
    },

    /**
     * The response time.
     *
     * @property time
     * @type number
     * @readonly
     */
    time: {
        get() { return this.tsData ? this.tsData - this.tsSubmit : undefined; }
    },

    /**
     * The abort timestamp.
     *
     * @property tsAbort
     * @type number
     * @readonly
     */
    tsAbort: {
        set(val) { return this.tsAbort || val; },
        validate(val) { return !XP.isInt(val, true) && 'number'; }
    },

    /**
     * The data timestamp.
     *
     * @property tsData
     * @type number
     * @readonly
     */
    tsData: {
        set(val) { return this.tsData || val; },
        validate(val) { return !XP.isInt(val, true) && 'number'; }
    },

    /**
     * The response timestamp.
     *
     * @property tsResponse
     * @type number
     * @readonly
     */
    tsResponse: {
        set(val) { return this.tsResponse || val; },
        validate(val) { return !XP.isInt(val, true) && 'number'; }
    },

    /**
     * The submit timestamp.
     *
     * @property tsSubmit
     * @type number
     * @readonly
     */
    tsSubmit: {
        set(val) { return this.tsSubmit || val; },
        validate(val) { return !XP.isInt(val, true) && 'number'; }
    },

    /**
     * The request's url.
     *
     * @property url
     * @type string
     */
    url: {
        set(val) { return this.url || val; },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    },

    /*********************************************************************/

    // HANDLER
    _handleDownload(event) {

        // Pushing
        if (http && this.parser !== 'buffer') { this.chunks.push(event); }

        // Emitting
        this.emit('download', this.downLoaded += http ? XPBuffer.byteLength(event) : event.loaded, this.downTotal);
    },

    // HANDLER
    _handleEnd(resolver, event) {

        // Let
        let data  = this.error || this.parser === 'buffer' ? null : (http ? this.chunks.join('') : event.target.response),
            state = this.error ? 'failed' : 'complete';

        // Parsing
        if (this.parser === 'json' && !this.error) { data = XP.toDefined(XP.parseJSON(data.toString())); }

        // Setting
        this.data   = data;
        this.state  = state;
        this.tsData = Date.now();

        // Emitting
        this.emit(this.error ? 'error' : 'data', this.error || data);

        // Resolving
        if (this.parser !== 'buffer') { resolver(this.error, data); }
    },

    // HANDLER
    _handleError(resolver, error) {

        // Setting
        this.error = XP.error(0, http ? error.message : 'Request not sent.');
        this.state = 'failed';

        // Emitting
        this.emit('error', this.error);

        // Resolving
        resolver(this.error, null);
    },

    // HANDLER
    _handleResponse(resolver, event) {

        // Checking
        if (XP.isDefined(this.statusCode)) { return; }

        // Setting
        this.response      = http ? event : event.target;
        this.statusCode    = this.response[http ? 'statusCode' : 'status'] || 502;
        this.statusMessage = this.response[http ? 'statusMessage' : 'statusText'] || codes[this.statusCode] || null;
        this.error         = this.statusCode >= 400 ? XP.error(this.statusCode, this.statusMessage) : null;
        this.downTotal     = XP.toDefined(XP.toFinite(http ? this.response.headers['content-length'] : this.response.getResponseHeader('Content-Length')));
        this.tsResponse    = Date.now();

        // Listening
        if (http) { this.response.on('data', this._handleDownload); }
        if (http) { this.response.on('end', this._handleEnd); }

        // Emitting
        this.emit('response', this.response);

        // Resolving
        if (this.parser === 'buffer') { resolver(this.error, this.error ? null : this.response); }
    }
});
