/*jslint browser: true, devel: true, node: true, ass: true, nomen: true, unparam: true, indent: 4 */

/**
 * @license
 * Copyright (c) 2017 The expand.js authors. All rights reserved.
 * This code may only be used under the BSD style license found at https://expandjs.github.io/LICENSE.txt
 * The complete set of authors may be found at https://expandjs.github.io/AUTHORS.txt
 * The complete set of contributors may be found at https://expandjs.github.io/CONTRIBUTORS.txt
 */

// Const
const env     = typeof window !== "undefined" ? window : global,
    stream    = typeof window !== "undefined" ? null : require('stream'),
    http      = typeof window !== "undefined" ? null : require('http'),
    https     = typeof window !== "undefined" ? null : require('https'),
    codes     = typeof window !== "undefined" ? {} : http.STATUS_CODES,
    location  = env.location || {},
    XP        = env.XP || require('expandjs'),
    XPBuffer  = env.XPBuffer || require('xp-buffer'),
    XPEmitter = env.XPEmitter || require('xp-emitter');

/*********************************************************************/

/**
 * A class used to provide uploading functionality.
 *
 * @class XPUploader
 * @extends XPEmitter /bower_components/xp-emitter/lib/index.js
 * @since 1.0.0
 * @description A class used to provide uploading functionality
 * @keywords nodejs, expandjs
 * @source https://github.com/expandjs/xp-uploader/blob/master/lib/index.js
 */
module.exports = global.XPUploader = new XP.Class('XPUploader', {

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
     * Emitted when the uploader's state changes.
     *
     * @event state
     * @param {string} state
     */

    /**
     * Emitted when the upload has started.
     *
     * @event submit
     * @param {File} file
     */

    /**
     * Emitted when a chunk of data is uploaded.
     *
     * @event upload
     * @param {number} loaded
     * @param {number} total
     */

    /*********************************************************************/

    /**
     * @constructs
     * @param {Object | string} options The uploader's url or options
     *   @param {Object} [options.headers] An object containing uploader headers
     *   @param {string} [options.hostname] The uploader's hostname, usable in alternative to url
     *   @param {number} [options.keepAlive = 0] How often to submit TCP KeepAlive packets over sockets being kept alive
     *   @param {string} [options.method = "POST"] A string specifying the HTTP upload method
     *   @param {string} [options.parser = "json"] The type of data expected back from the server
     *   @param {string} [options.path] The uploader's path, usable in alternative to url
     *   @param {number} [options.port] The uploader's port, usable in alternative to url
     *   @param {number} [options.protocol = "http:"] The uploader's protocol, usable in alternative to url
     *   @param {string} [options.url] The uploader's url
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
                this.upLoaded   = 0;
                this.chunks     = [];
                this.state      = 'idle';
                this.options    = options;
                this.headers    = this.options.headers || {};
                this.hostname   = this.options.hostname || location.hostname || '';
                this.keepAlive  = this.options.keepAlive || 0;
                this.method     = this.options.method || 'POST';
                this.parser     = this.options.parser || 'json';
                this.path       = this.options.path || '';
                this.port       = this.options.port || (!this.options.hostname && location.port) || null;
                this.protocol   = this.options.protocol || (!this.options.hostname && location.protocol) || 'http:';
                this.url        = XP.toURL({protocol: this.protocol, hostname: this.hostname, port: this.port, pathname: this.pathname, search: this.search});

                // Binding
                this._handleDownload = this._handleDownload.bind(this);
                this._handleEnd      = this._handleEnd.bind(this, resolver);
                this._handleError    = this._handleError.bind(this, resolver);
                this._handlePipe     = this._handlePipe.bind(this);
                this._handleResponse = this._handleResponse.bind(this, resolver);
                this._handleUpload   = this._handleUpload.bind(this);

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
     * Aborts the upload.
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
     * Starts the upload of `source`.
     *
     * @method submit
     * @param {Object} source
     * @param {number} [offset = 0]
     * @param {Function} [callback]
     * @returns {Object}
     */
    submit: {
        callback: true,
        value(source, offset, callback) {

            // Checking
            if (!XP.isInstance(source, http ? stream.Readable : env.File)) { callback(XP.error(500, `"data" must be ${http ? `Readable` : `File`}.`)); return; }
            if (!XP.isVoid(offset) && !XP.isInt(offset, true)) { callback(XP.error(500, `"offset" must be number.`)); return; }

            // Preventing
            if (this.tsSubmit) { return; }

            // Callback
            this.ready(callback);

            // Setting
            this.upTotal  = http ? null : source.size;
            this.upLoaded = http ? 0 : offset || 0;

            // CASE: server
            if (http) {

                // Adapting
                let request = this.adaptee = (this.protocol.startsWith('https') ? https : http).request({
                    hostname: this.hostname,
                    keepAlive: this.keepAlive > 0,
                    keepAliveMsecs: this.keepAlive,
                    method: this.method,
                    path: this.path,
                    port: this.port,
                    protocol: this.protocol,
                    withCredentials: false
                });

                // Headers
                Object.keys(this.headers).forEach(key => XP.isInput(this.headers[key], true) && request.setHeader(key, this.headers[key]));
                request.setHeader('Content-Type', source.headers['content-type'] || 'application/octet-stream');
                request.setHeader('Content-Range', source.headers['content-range']);

                // Listening
                request.on('error',    this._handleError);
                request.on('pipe',     this._handlePipe);
                request.on('response', this._handleResponse);
            }

            // CASE: browser
            if (!http) {

                // Adapting
                let xhr = this.adaptee = new XMLHttpRequest();

                // Opening
                xhr.open(this.method, this.url, true);

                // Headers
                Object.keys(this.headers).forEach(key => XP.isInput(this.headers[key], true) && xhr.setRequestHeader(key, this.headers[key]));
                xhr.setRequestHeader('Content-Type', source.type || 'application/octet-stream');
                xhr.setRequestHeader('Content-Range', `bytes ${this.upLoaded}-${source.size - 1}/${source.size}`);

                // Listening
                xhr.addEventListener('progress', this._handleResponse);
                xhr.addEventListener('progress', this._handleDownload);
                xhr.addEventListener('load', this._handleResponse);
                xhr.addEventListener('load', this._handleEnd);
                xhr.addEventListener('error', this._handleError);
                xhr.upload.addEventListener('progress', this._handleUpload);
                xhr.upload.addEventListener('error', this._handleError);
            }

            // Setting
            this.state    = 'pending';
            this.tsSubmit = Date.now();

            // Sending
            if (http) { source.pipe(this.adaptee); } else { this.adaptee.send(source); }

            // Emitting
            this.emit('submit', source);

            // Chaining
            return this;
        }
    },

    /*********************************************************************/

    /**
     * The uploader's instance.
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
     * An object containing uploader headers.
     *
     * @property headers
     * @type Object
     */
    headers: {
        set(val) { return this.headers || (XP.isObject(val) && XP.cloneDeep(val)); },
        validate(val) { return !XP.isObject(val) && 'Object'; }
    },

    /**
     * The uploader's host.
     *
     * @property host
     * @type string
     * @readonly
     */
    host: {
        get() { return `${this.hostname || ``}${this.port ? `:${this.port}` : ``}`; }
    },

    /**
     * The uploader's hostname.
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
     * A string specifying the HTTP upload method.
     *
     * @property method
     * @type string
     * @default "POST"
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
     * The uploader's path.
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
     * The uploader's pathname.
     *
     * @property pathname
     * @type string
     */
    pathname: {
        set(val) { return XP.isDefined(this.pathname) ? this.pathname : val; },
        validate(val) { return !XP.isString(val) && 'string'; }
    },

    /**
     * The uploader's port.
     *
     * @property port
     * @type number
     */
    port: {
        set(val) { return XP.isDefined(this.port) ? this.port : val; },
        validate(val) { return !XP.isNull(val) && !XP.isNumeric(val, true) && 'number'; }
    },

    /**
     * The uploader's protocol.
     *
     * @property protocol
     * @type string
     */
    protocol: {
        set(val) { return this.protocol || val; },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    },

    /**
     * The uploader's query.
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
     * The uploader's response.
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
     * The uploader's search.
     *
     * @property search
     * @type string
     */
    search: {
        get() { return this.query ? `?${this.query}` : ``; }
    },

    /**
     * The uploader's state.
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
     * The list of possible uploader's states.
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
     * The uploader's status code.
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
     * The uploader's status code.
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
     * The uploaded bytes.
     *
     * @property upLoaded
     * @type Number
     * @readonly
     */
    upLoaded: {
        set(val) { return val; },
        validate(val) { return !XP.isInt(val, true) && 'number'; }
    },

    /**
     * The upload ratio.
     *
     * @property upRatio
     * @type Number
     * @readonly
     */
    upRatio: {
        get() { return XP.isVoid(this.downTotal) ? null : (this.downTotal ? this.downLoaded / this.downTotal : 1); }
    },

    /**
     * The upload total.
     *
     * @property upTotal
     * @type Number
     * @readonly
     */
    upTotal: {
        set(val) { return XP.isDefined(this.downTotal) ? this.downTotal : val; },
        validate(val) { return !XP.isNull(val) && !XP.isInt(val, true) && 'number'; }
    },

    /**
     * The uploader's url.
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
    _handlePipe(source) {

        // Listening
        source.on('data', this._handleUpload);
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
    },

    // HANDLER
    _handleUpload(event) {

        // Emitting
        this.emit('upload', this.upLoaded += http ? XPBuffer.byteLength(event) : event.loaded, this.upTotal);
    }
});
