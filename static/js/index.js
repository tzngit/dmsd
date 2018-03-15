var app = angular.module('dbms', ['ui.bootstrap']);

app.filter("trustUrl", ['$sce', function ($sce) {
    return function (recordingUrl) {
        return $sce.trustAsResourceUrl(recordingUrl);
    };
}]);

app.controller('BodyController', function($scope, $http) {
    $scope.song_records = [];

    $scope.parse = function() {
        $http.post("/parse", {URL:$scope.site_url}).success(function(data) {
            console.log("parse ret:", data)
            $scope.song_records = data;
        }).error(function(data) {
            console.log("parse error\n", data)
        });
    };
});
