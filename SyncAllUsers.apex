public class SyncAllUsers {
    @future(callout=true)
    public static void sendUsersToHeroku() {
        List<User> users = [
            SELECT Id, Username, Email, FirstName, LastName, CreatedDate, Phone, Pin__c
            FROM User
            WHERE IsActive = true
        ];

        List<Map<String, Object>> payload = new List<Map<String, Object>>();
        for (User u : users) {
            payload.add(new Map<String, Object>{
            'Id' => u.Id,
            'Username' => u.Username,
            'Email' => u.Email,
            'FirstName' => u.FirstName,
            'LastName' => u.LastName,
            'Phone' => u.Phone,
            'Pin__c' => u.Pin__c,
            'CreatedDate' => String.valueOf(u.CreatedDate)
            });

        }

        HttpRequest req = new HttpRequest();
        req.setEndpoint('callout:HerokuAPI/sync-users');
        req.setMethod('POST');
        req.setHeader('Content-Type', 'application/json');
        req.setBody(JSON.serialize(payload));

        Http http = new Http();
        try {
            HttpResponse res = http.send(req);
            System.debug('Response: ' + res.getStatus() + ' ' + res.getBody());
        } catch (Exception e) {
            System.debug('Callout failed: ' + e.getMessage());
        }
    }
}