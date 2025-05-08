// Trigger: UserSyncTrigger.trigger
trigger UserSyncTrigger on User (after insert, after update) {
    // Only run the sync once per transaction (skip duplicate calls in bulk ops)
    if (!System.isBatch()) {
        SyncAllUsers.sendUsersToHeroku();
    }
}