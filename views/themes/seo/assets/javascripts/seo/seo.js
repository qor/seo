(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define(['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function($) {
    "use script";

    var NAMESPACE = 'qor.seo';
    var EVENT_ENABLE = 'enable.' + NAMESPACE;
    var EVENT_DISABLE = 'disable.' + NAMESPACE;
    var EVENT_CLICK = 'click.' + NAMESPACE;
    var EVENT_BLUR = 'blur.' + NAMESPACE;
    var CLASS_SUBMIT = ".qor-seo-submit";
    var CLASS_ADD_TAGS_NAME = ".qor-seo-tag";
    var CLASS_TAGS_INPUT_NAME = ".qor-seo-input-field";
    var CLASS_DEFAULT_INPUT = ".qor-seo__defaults-input";
    var CLASS_SETTINGS = ".qor-seo__settings";

    function QorSeo(element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorSeo.DEFAULTS, $.isPlainObject(options) && options);
        this.focusedInputID = "";
        this.init();
    }

    QorSeo.prototype={

        constructor: QorSeo,

        init: function() {
            var $element = this.$element;

            this.$submit = $element.find(CLASS_SUBMIT);
            this.$addTgas = $element.find(CLASS_ADD_TAGS_NAME);
            this.$tagInputs = $element.find(CLASS_TAGS_INPUT_NAME);
            this.$settingInput = $element.find(CLASS_DEFAULT_INPUT);
            this.bind();
        },

        bind: function () {
            this.$submit.on(EVENT_CLICK, $.proxy(this.submitSeo, this));
            this.$tagInputs.on('click keyup', $.proxy(this.tagInputsFocus, this));
            this.$tagInputs.on(EVENT_BLUR, $.proxy(this.tagInputsBlur, this));
            this.$settingInput.on(EVENT_CLICK, $.proxy(this.toggleDefault, this));
            this.$addTgas.on(EVENT_CLICK, $.proxy(this.addTags, this));
        },

        unbind: function () {
            this.$submit.off(EVENT_CLICK, $.proxy(this.submitSeo, this));
            this.$tagInputs.off('click keyup', $.proxy(this.tagInputsFocus, this));
            this.$tagInputs.off(EVENT_BLUR, $.proxy(this.tagInputsBlur, this));
            this.$settingInput.off(EVENT_CLICK, $.proxy(this.toggleDefault, this));
            this.$addTgas.off(EVENT_CLICK, $.proxy(this.addTags, this));
        },

        toggleDefault: function () {
            var isChecked = this.$settingInput.is(':checked'),
                $settings = $(CLASS_SETTINGS);

            isChecked ? $settings.hide() : $settings.show();
        },

        tagInputsFocus: function () {
            this.$addTgas.addClass('focus');
            var $focusedInput = $(document.activeElement);

            this.focusedInputID = $focusedInput.prop("id");
            this.focusedInputStart = $focusedInput[0].selectionStart;
            this.focusedInputEnd = $focusedInput[0].selectionEnd;
            this.focusedInputVal = $focusedInput.val();
        },

        tagInputsBlur: function () {
            this.$addTgas.removeClass('focus');
            this.$focusedInputID = false;
        },

        addTags: function (e) {
            if (!this.focusedInputID){
                return;
            }

            var newVal = "";
            var startString = this.focusedInputVal.substring(0,this.focusedInputStart);
            var endString = this.focusedInputVal.substring(this.focusedInputEnd,this.focusedInputVal.length);
            var tagVal = "{{"+$(e.currentTarget).data("tagValue")+"}}";

            newVal = startString + tagVal + endString;
            $("#"+this.focusedInputID).val(newVal).focus();
        },

        submitSeo: function() {
            var $element = this.$element,
                $form = $element.find(".qor-form");


            // new FormData(form)
            // "GlobalSetting": { "SiteName" : "Qor", "BrandName" : "ThePlant" }

            $.ajax({
                type: "POST",
                url: $form.attr("action"),
                data: $form.serialize(),
                success: function () {
                    window.onbeforeunload = null;
                    $.fn.qorSlideoutBeforeHide = null;
                    $('.qor-alert--success').show().addClass('');
                    setTimeout(function () {
                        $('.qor-alert--success').hide();
                      }, 5000);
                },
                error: function () {
                    $('.qor-alert--error').show();
                }
            });
            return false;
        },

        destroy: function () {
            this.unbind();
            this.$element.removeData(NAMESPACE);
        }
    };

    QorSeo.DEFAULTS = {};

    QorSeo.plugin = function (options) {
        return this.each(function () {
            var $this = $(this);
            var data = $this.data(NAMESPACE);
            var fn;

            if (!data) {
                if (/destroy/.test(options)) {
                    return;
                }

                $this.data(NAMESPACE, (data = new QorSeo(this, options)));
            }

            if (typeof options === 'string' && $.isFunction(fn = data[options])) {
                fn.apply(data);
            }
        });
    };

    $(function () {
        var selector = '[data-toggle="qor.seo"]',
            options = {};

        $(document).
            on('click.qor.fixedAlert', '[data-dismiss="fixed-alert"]', function () {
                $(this).closest('.qor-alert').hide();
            }).
            on(EVENT_DISABLE, function (e) {
            QorSeo.plugin.call($(selector, e.target), 'destroy');
            }).
            on(EVENT_ENABLE, function (e) {
            QorSeo.plugin.call($(selector, e.target), options);
            }).
            triggerHandler(EVENT_ENABLE);

    });

    return QorSeo;
});
