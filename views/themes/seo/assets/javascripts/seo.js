$(".qor-seo-submit").click(function(){
  var $wrap = $(this).parents(".qor-seo");
  var fieldName = $wrap.find(".qor-seo-field").attr("name");
  var titleValue = $wrap.find(".qor-seo-title-field").val();
  var tagsValue = $wrap.find(".qor-seo-tags-field").val();
  var descriptionValue = $wrap.find(".qor-seo-description-field").val();
  var data = {};
  data[fieldName] = '{ "Title" : "' + titleValue + '", "Description" : "' + descriptionValue + '", "Tags" : "' + tagsValue + '"}';
  console.info(data);
  $.ajax({
    type: "POST",
    url: "/admin/seo/1",
    data: data,
    success: function () {
      alert("Save");
    },
    error: function (data) {
      alert("Can't save");
    }
  });
  return false;
});
